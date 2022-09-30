package recorder

import (
	"fmt"
	"github.com/dunkbing/meeting-bot/pkg/display"
	pipeline2 "github.com/dunkbing/meeting-bot/pkg/pipeline"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Recorder struct {
	ID string

	display  *display.Display
	pipeline *pipeline2.Pipeline
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
	r.pipeline, err = r.createPipeline(pipeline2.Options{
		Width:          1024,
		Height:         768,
		Depth:          24,
		Framerate:      30,
		AudioBitrate:   128,
		AudioFrequency: 44100,
		VideoBitrate:   4500,
		Profile:        "main",
	})
	if err != nil {
		logrus.Errorln("Error building pipeline", err)
		return err
	}

	go func() {
		r.mu.Lock()
		defer r.mu.Unlock()
	}()

	// run pipeline
	err = r.pipeline.Run()
	if err != nil {
		return err
	}

	return nil
}

func (r *Recorder) createPipeline(options pipeline2.Options) (*pipeline2.Pipeline, error) {
	return pipeline2.NewFilePipeline(r.filename, options)
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
