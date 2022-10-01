package display

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/dunkbing/meeting-bot/pkg/bot"
	"github.com/dunkbing/meeting-bot/pkg/config"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
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
	_cfg, _ := config.Get()

	width, height := 1440, 900
	if err := d.launchXvfb(_cfg.Display, width, height, 24); err != nil {
		return nil, err
	}
	if err := d.launchChrome(_cfg.Display, width, height); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Display) launchXvfb(display string, width, height, depth int) error {
	dims := fmt.Sprintf("%dx%dx%d", width, height, depth)
	logrus.Infoln("Launching xvfb.", "dims:", dims)
	xvfb := exec.Command("Xvfb", display, "-screen", "0", dims, "-ac", "-nolisten", "tcp")
	if err := xvfb.Start(); err != nil {
		logrus.Errorln("Error launching xvfb:", err.Error())
		return err
	}
	d.xvfb = xvfb
	return nil
}

func (d *Display) launchChrome(display string, width, height int) error {
	logrus.Infoln("Launching chrome")
	cfg, _ := config.Get()
	type_ := bot.GetMeetingType(cfg.MeetingUrl)

	if type_ == bot.InvalidType {
		return nil
	}

	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.NoSandbox,

		// puppeteer default behavior
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("excludeSwitches", "enable-automation"),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-features", "site-per-process,TranslateUI,BlinkGenPropertyTrees"),
		//chromedp.Flag("disable-software-rasterizer", true),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("use-mock-keychain", true),

		// custom args
		chromedp.Flag("kiosk", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
		chromedp.Flag("window-position", "0,0"),
		chromedp.Flag("window-size", fmt.Sprintf("%d,%d", width, height)),
		chromedp.Flag("display", display),
	}

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	d.chromeCancel = cancel

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			args := make([]string, 0, len(ev.Args))
			for _, arg := range ev.Args {
				var val interface{}
				err := json.Unmarshal(arg.Value, &val)
				if err != nil {
					continue
				}
				msg := fmt.Sprint(val)
				args = append(args, msg)
				switch msg {
				default:
				}
			}
			logrus.Debugln(fmt.Sprintf("chrome console %s", ev.Type.String()), "msg", strings.Join(args, " "))
		}
	})
	err := chromedp.Run(ctx, chromedp.Navigate("https://www.youtube.com/watch?v=WgnFgUq_KFw"))
	return err
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
