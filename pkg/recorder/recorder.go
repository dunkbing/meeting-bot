package recorder

import (
	"errors"
	"fmt"
	"github.com/dunkbing/meeting-bot/pkg/bot"
	"github.com/sirupsen/logrus"
	"sync"
	"time"

	"github.com/dunkbing/meeting-bot/pkg/config"
	"github.com/dunkbing/meeting-bot/pkg/display"
	"github.com/dunkbing/meeting-bot/pkg/gst"
)

type RtmpResult struct {
	StreamUrl string
	Duration  int64
}

type FileResult struct {
	DownloadUrl string
	Duration    int64
}

type RecordingInfo struct {
	Id    string
	Error string
	Rtmp  []*RtmpResult
	File  *FileResult
}

type RtmpOutput struct {
	Urls []string `protobuf:"bytes,1,rep,name=urls,proto3" json:"urls,omitempty"`
}

type StartRecordingRequestRtmp struct {
	Rtmp RtmpOutput
	Urls []string
}

type RecordingOptions struct {
	Width          int32  // default 1920
	Height         int32  // default 1080
	Depth          int32  // default 24
	Framerate      int32  // default 30
	AudioBitrate   int32  // default 128
	AudioFrequency int32  // default 44100
	VideoBitrate   int32  // default 4500
	Profile        string // baseline, main, or high. default main
}

type RequestType string

const file RequestType = "FILE"
const rtmp RequestType = "RTMP"

type StartRecordingRequest struct {
	Output  *StartRecordingRequestRtmp
	Type    RequestType
	Options *RecordingOptions
}

type Recorder struct {
	ID string

	req      *StartRecordingRequest
	display  *display.Display
	pipeline *gst.Pipeline
	abort    chan struct{}

	filepath string

	mu        sync.Mutex
	result    *RecordingInfo
	startedAt map[string]time.Time
}

func NewRecorder(recordingID string) *Recorder {
	conf, _ := config.GetConfig()
	return &Recorder{
		ID: recordingID,

		abort: make(chan struct{}),

		filepath: fmt.Sprintf("%s/%s", conf.FileOutput.FileDir, conf.FileOutput.FileName),

		result: &RecordingInfo{
			Id: recordingID,
		},
		startedAt: make(map[string]time.Time),
	}
}

// Run blocks until completion
func (r *Recorder) Run() *RecordingInfo {
	var err error
	conf, _ := config.GetConfig()

	// launch display
	r.display, err = display.Launch()
	b := bot.New()
	err = b.Start()
	if err != nil {
		logrus.Error("error launching display", err)
		r.result.Error = err.Error()
		return r.result
	}

	// create gst
	r.pipeline, err = r.createPipeline(r.req)
	if err != nil {
		logrus.Error("error building gst: ", err)
		r.result.Error = err.Error()
		return r.result
	}

	var startedAt time.Time
	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
	}()

	// run gst
	err = r.pipeline.Run()
	if err != nil {
		logrus.Error("error running gst", err)
		r.result.Error = err.Error()
		return r.result
	}

	//t := r.req.Type
	switch file {
	case rtmp:
		for url, startTime := range r.startedAt {
			r.result.Rtmp = append(r.result.Rtmp, &RtmpResult{
				StreamUrl: url,
				Duration:  time.Since(startTime).Milliseconds() / 1000,
			})
		}
	case file:
		r.result.File = &FileResult{
			Duration: time.Since(startedAt).Milliseconds() / 1000,
		}
		if conf.FileOutput.S3 != nil {
			if err = r.uploadS3(); err != nil {
				r.result.Error = err.Error()
				return r.result
			}
			r.result.File.DownloadUrl = fmt.Sprintf("s3://%s/%s", conf.FileOutput.S3.Bucket, r.filepath)
		} else if conf.FileOutput.AzBlob != nil {
			if err = r.uploadAzure(); err != nil {
				r.result.Error = err.Error()
				return r.result
			}
			r.result.File.DownloadUrl = fmt.Sprintf(
				"https://%s.blob.core.windows.net/%s/%s",
				conf.FileOutput.AzBlob.AccountName,
				conf.FileOutput.AzBlob.ContainerName,
				r.filepath,
			)
		} else if conf.FileOutput.GCPConfig != nil {
			if err = r.uploadGCP(); err != nil {
				r.result.Error = err.Error()
				return r.result
			}
			r.result.File.DownloadUrl = fmt.Sprintf("gs://%s/%s", conf.FileOutput.GCPConfig.Bucket, r.filepath)
		}
	}

	return r.result
}

func (r *Recorder) createPipeline(req *StartRecordingRequest) (*gst.Pipeline, error) {
	output := file
	switch output {
	case rtmp:
		return gst.NewRtmpPipeline(req.Output.Rtmp.Urls)
	case file:
		return gst.NewFilePipeline(r.filepath)
	}
	return nil, errors.New("no output")
}

func (r *Recorder) AddOutput(url string) error {
	logrus.Debug("add output", "url", url)
	if r.pipeline == nil {
		return gst.ErrPipelineNotFound
	}

	if err := r.pipeline.AddOutput(url); err != nil {
		return err
	}
	startedAt := time.Now()

	r.mu.Lock()
	r.startedAt[url] = startedAt
	r.mu.Unlock()

	return nil
}

func (r *Recorder) RemoveOutput(url string) error {
	logrus.Debug("remove output", "url", url)
	if r.pipeline == nil {
		return gst.ErrPipelineNotFound
	}

	if err := r.pipeline.RemoveOutput(url); err != nil {
		return err
	}
	endedAt := time.Now()

	r.mu.Lock()
	if startedAt, ok := r.startedAt[url]; ok {
		r.result.Rtmp = append(r.result.Rtmp, &RtmpResult{
			StreamUrl: url,
			Duration:  endedAt.Sub(startedAt).Milliseconds() / 1000,
		})
		delete(r.startedAt, url)
	}
	r.mu.Unlock()

	return nil
}

func (r *Recorder) Stop() {
	select {
	case <-r.abort:
		return
	default:
		close(r.abort)
		if p := r.pipeline; p != nil {
			p.Close()
		}
	}
}

// Close should only be called after gst completes
func (r *Recorder) Close() {
	if r.display != nil {
		r.display.Close()
	}
}
