//go:build test
// +build test

package gst

import (
	"errors"
	"github.com/dunkbing/meeting-bot/pkg/config"
	"time"
)

type Pipeline struct {
	isStream  bool
	startedAt time.Time
	kill      chan struct{}
}

func NewRtmpPipeline(urls []string, options *config.Config) (*Pipeline, error) {
	return &Pipeline{
		isStream: true,
		kill:     make(chan struct{}, 1),
	}, nil
}

func NewFilePipeline(filename string, options *config.Config) (*Pipeline, error) {
	return &Pipeline{
		isStream: false,
		kill:     make(chan struct{}, 1),
	}, nil
}

func (p *Pipeline) Run() error {
	p.startedAt = time.Now()
	select {
	case <-time.After(time.Second * 3):
	case <-p.kill:
	}
	return nil
}

func (p *Pipeline) GetStartTime() time.Time {
	return p.startedAt
}

func (p *Pipeline) AddOutput(url string) error {
	if !p.isStream {
		return errors.New("cannot add rtmp output to file recording")
	}
	return nil
}

func (p *Pipeline) RemoveOutput(url string) error {
	if !p.isStream {
		return errors.New("cannot remove rtmp output from file recording")
	}
	return nil
}

func (p *Pipeline) Abort() {
	p.kill <- struct{}{}
}

func (p *Pipeline) Close() {
	p.kill <- struct{}{}
}
