package display

import (
	"context"
	"fmt"
	"github.com/dunkbing/meeting-bot/bot"
	"github.com/dunkbing/meeting-bot/config"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
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

	width, height := 1440, 900
	if err := d.launchXvfb(":0", width, height, 24); err != nil {
		return nil, err
	}
	if err := d.launchChrome(width, height); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Display) launchXvfb(display string, width, height, depth int) error {
	dims := fmt.Sprintf("%dx%dx%d", width, height, depth)
	logrus.Infoln("Launching xvfb", "dims", dims)
	xvfb := exec.Command("Xvfb", display, "-screen", "0", dims, "-ac", "-nolisten", "tcp")
	if err := xvfb.Start(); err != nil {
		return err
	}
	d.xvfb = xvfb
	return nil
}

func (d *Display) launchChrome(width, height int) error {
	logrus.Infoln("Launching chrome")
	cfg, _ := config.Get()
	type_ := bot.GetMeetingType(cfg.MeetingUrl)

	if type_ == bot.InvalidType {
		return nil
	}

	u := launcher.New().
		Set("user-data-dir", "chrome-data").
		//Set("headless").
		Set("incognito").
		Set("use-fake-ui-for-media-stream").
		Set("autoplay-policy", "no-user-gesture-required").
		Set("disable-gpu").
		Set("disable-software-rasterizer").
		Set("disable-dev-shm-usage").
		Set("bwsi").
		Set("no-first-run").
		Set("no-sandbox").
		Set("window-position", "0,0").
		Set("window-size", fmt.Sprintf("%d,%d", width, height)).
		//Set("start-maximized").
		Set("disable-blink-features", "AutomationControlled").
		Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36").
		MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect().NoDefaultDevice()
	browser.MustPage("https://www.youtube.com/watch?v=WgnFgUq_KFw")

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
			logrus.Errorln("failed to kill xvfb", err)
		}
		d.xvfb = nil
	}
}
