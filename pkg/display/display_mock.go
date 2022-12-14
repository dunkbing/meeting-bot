//go:build test
// +build test

package display

import (
	"github.com/dunkbing/meeting-bot/pkg/config"
)

type Display struct {
	startChan chan struct{}
	endChan   chan struct{}
}

func Launch(conf *config.Config, url string, isTemplate bool) (*Display, error) {
	startChan := make(chan struct{})
	close(startChan)

	return &Display{
		startChan: startChan,
		endChan:   make(chan struct{}),
	}, nil
}

func (d *Display) RoomStarted() chan struct{} {
	return d.startChan
}

func (d *Display) RoomEnded() chan struct{} {
	return d.endChan
}

func (d *Display) Close() {
	close(d.endChan)
}
