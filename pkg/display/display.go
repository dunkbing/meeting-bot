//go:build !test
// +build !test

package display

import (
	"context"
	"fmt"
	"github.com/dunkbing/meeting-bot/pkg/config"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

const (
	startRecording = "START_RECORDING"
	endRecording   = "END_RECORDING"
)

type Display struct {
	xvfb         *exec.Cmd
	chromeCancel context.CancelFunc
	startChan    chan struct{}
	endChan      chan struct{}
}

func Launch() (*Display, error) {
	d := &Display{
		startChan: make(chan struct{}),
		endChan:   make(chan struct{}),
	}
	conf, _ := config.GetConfig()

	if err := d.launchXvfb(conf.Screen); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Display) launchXvfb(screenConf config.ScreenConfig) error {
	dims := fmt.Sprintf("%dx%dx%d", screenConf.Width, screenConf.Height, screenConf.Depth)
	logrus.Debug("launching xvfb", "dims", dims)
	xvfb := exec.Command("Xvfb", screenConf.Display, "-screen", "0", dims, "-ac", "-nolisten", "tcp")
	if err := xvfb.Start(); err != nil {
		return err
	}
	d.xvfb = xvfb
	return nil
}

func (d *Display) RoomStarted() chan struct{} {
	return d.startChan
}

func (d *Display) RoomEnded() chan struct{} {
	return d.endChan
}

func (d *Display) Close() {
	if d.chromeCancel != nil {
		d.chromeCancel()
		d.chromeCancel = nil
	}

	if d.xvfb != nil {
		err := d.xvfb.Process.Signal(os.Interrupt)
		if err != nil {
			logrus.Error("failed to kill xvfb", err)
		}
		d.xvfb = nil
	}
}
