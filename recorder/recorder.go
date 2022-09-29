package recorder

import (
	"fmt"
	"github.com/dunkbing/meeting-bot/display"
	"github.com/dunkbing/meeting-bot/pipeline"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Recorder struct {
	ID string

	display  *display.Display
	pipeline *pipeline.Pipeline
	abort    chan struct{}

	isTemplate bool
	url        string
	filename   string
	filepath   string

	// result info
	mu        sync.Mutex
	startedAt map[string]time.Time
}

func NewRecorder(recordingID string) *Recorder {
	return &Recorder{
		ID:        recordingID,
		filename:  "test.mp4",
		filepath:  ".",
		abort:     make(chan struct{}),
		startedAt: make(map[string]time.Time),
	}
}

// Run blocks until completion
func (r *Recorder) Run() error {
	logrus.Println("Recorder starts")
	var err error

	// launch display
	r.display, err = display.Launch()
	if err != nil {
		fmt.Println("error launching display", err)
		return err
	}

	// create pipeline
	r.pipeline, err = r.createPipeline(pipeline.Options{
		Width:          1024,
		Height:         768,
		Depth:          24,
		Framerate:      30,
		AudioBitrate:   200,
		AudioFrequency: 44100,
		VideoBitrate:   4500,
		Profile:        "main",
	})
	if err != nil {
		logrus.Errorln("Error building pipeline", err)
		return err
	}

	// if using template, listen for START_RECORDING and END_RECORDING messages
	if r.isTemplate {
		logrus.Infoln("Waiting for room to start")
		select {
		case <-r.display.RoomStarted():
			logrus.Infoln("Room started")
		case <-r.abort:
			r.pipeline.Abort()
			logrus.Infoln("Recording aborted while waiting for room")
			return nil
		}

		// stop on END_RECORDING console log
		go func(d *display.Display) {
			<-d.RoomEnded()
			r.Stop()
		}(r.display)
	}

	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
	}()

	// run pipeline
	err = r.pipeline.Run()
	if err != nil {
		logrus.Errorln("error running pipeline", err)
		return err
	}

	return nil
}

func (r *Recorder) createPipeline(options pipeline.Options) (*pipeline.Pipeline, error) {
	return pipeline.NewFilePipeline(r.filename, options)
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
