package pipeline

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"
	"strings"
	"sync"
	"time"
)

func requireLink(src, sink *gst.Pad) error {
	if linkReturn := src.Link(sink); linkReturn != gst.PadLinkOK {
		return fmt.Errorf("pad link: %s", linkReturn.String())
	}
	return nil
}

type Pipeline struct {
	mu sync.Mutex

	pipeline *gst.Pipeline
	loop     *glib.MainLoop

	output  *OutputBin
	removed map[string]bool

	started   chan struct{}
	startedAt time.Time
	closed    chan struct{}

	err error
}

var initialized = false

func NewFilePipeline(filename string, options Options) (*Pipeline, error) {
	if !initialized {
		gst.Init(nil)
		initialized = true
	}

	input, err := newInputBin(false, options)
	if err != nil {
		return nil, err
	}
	output, err := newFileOutputBin(filename)
	if err != nil {
		return nil, err
	}

	return newPipeline(input, output)
}

func newPipeline(input *InputBin, output *OutputBin) (*Pipeline, error) {
	// elements must be added to pipeline before linking
	pipeline, err := gst.NewPipeline("pipeline")
	if err != nil {
		return nil, err
	}

	// add bins to pipeline
	if err = pipeline.AddMany(input.bin.Element, output.bin.Element); err != nil {
		return nil, err
	}

	// link bin elements
	if err = input.Link(); err != nil {
		return nil, err
	}

	// link bins
	if err = input.bin.Link(output.bin.Element); err != nil {
		return nil, err
	}

	return &Pipeline{
		pipeline: pipeline,
		output:   output,
		removed:  make(map[string]bool),
		started:  make(chan struct{}),
		closed:   make(chan struct{}),
	}, nil
}

const pipelineSource = "pipeline"

func (p *Pipeline) Run() error {
	// add watch
	p.loop = glib.NewMainLoop(glib.MainContextDefault(), false)
	bus := p.pipeline.GetPipelineBus()
	bus.AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS:
			// EOS received - close and return
			logrus.Infoln("EOS received, stopping pipeline")
			_ = p.pipeline.BlockSetState(gst.StateNull)
			logrus.Infoln("Pipeline stopped")

			p.loop.Quit()
			return false
		case gst.MessageError:
			// handle error if possible, otherwise close and return
			gErr := msg.ParseError()
			p.err = gErr
			err, handled := p.handleError(gErr)
			if handled {
				logrus.Infoln("error handled", errors.New(gErr.Error()))
			} else {
				p.err = err
				p.loop.Quit()
				return false
			}
		case gst.MessageStateChanged:
			if msg.Source() == pipelineSource && p.startedAt.IsZero() {
				_, newState := msg.ParseStateChanged()
				if newState == gst.StatePlaying {
					p.startedAt = time.Now()
					close(p.started)
				}
			}
		default:
			logrus.Infoln("default", msg.Type(), "message", msg.String())
		}

		return true
	})

	// set state to playing (this does not start the pipeline)
	if err := p.pipeline.SetState(gst.StatePlaying); err != nil {
		return err
	}

	// run main loop
	p.loop.Run()
	return p.err
}

func (p *Pipeline) GetStartTime() time.Time {
	select {
	case <-p.started:
		return p.startedAt
	case <-p.closed:
		return p.startedAt
	}
}

func (p *Pipeline) Abort() {
	select {
	case <-p.closed:
		return
	default:
		close(p.closed)
	}
}

// Close waits for the pipeline to start before closing
func (p *Pipeline) Close() {
	select {
	case <-p.closed:
		return
	case <-p.started:
		close(p.closed)

		logrus.Debugln("Sending EOS to pipeline")
		p.pipeline.SendEvent(gst.NewEOSEvent())
	}
}

// Debug info comes in the following format:
// file.c(line): method_name (): /GstPipeline:pipeline/GstBin:bin_name/GstElement:element_name:\nError message
func parseDebugInfo(debug string) (element string, reason string, ok bool) {
	end := strings.Index(debug, ":\n")
	if end == -1 {
		return
	}
	start := strings.LastIndex(debug[:end], ":")
	if start == -1 {
		return
	}
	element = debug[start+1 : end]
	reason = debug[end+2:]
	if strings.HasPrefix(reason, GErrCouldNotConnect) {
		reason = GErrCouldNotConnect
	}
	ok = true
	return
}

// handleError returns true if the error has been handled, false if the pipeline should quit
func (p *Pipeline) handleError(gErr *gst.GError) (error, bool) {
	err := errors.New(gErr.Error())

	element, reason, ok := parseDebugInfo(gErr.DebugString())
	if !ok {
		logrus.Errorln("Failed to parse pipeline error", err, "debug", gErr.DebugString())
		return err, false
	}

	switch reason {
	case GErrNoURI, GErrCouldNotConnect:
		logrus.Errorln("Case gerrnouri", err)
		return err, true
	case GErrFailedToStart:
		// returned after an added rtmp sink failed to start
		// should be preceded by a GErrNoURI on the same sink
		handled := p.removed[element]
		if !handled {
			logrus.Errorln("Element failed to start", err)
		}
		return err, handled
	case GErrStreamingStopped:
		// returned by queue after rtmp sink could not connect
		// should be preceded by a GErrCouldNotConnect on associated sink
		handled := false
		if strings.HasPrefix(element, "queue_") {
			handled = p.removed[fmt.Sprint("sink_", element[6:])]
		}
		if !handled {
			logrus.Errorln("Streaming sink stopped", err)
		}
		return err, handled
	default:
		// input failure or file write failure. Fatal
		logrus.Errorln("Pipeline error", err, "debug", gErr.DebugString())
		return err, false
	}
}
