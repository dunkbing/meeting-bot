package main

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"log"
	"time"
)

func main() {
	u := launcher.New().
		//Bin("C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe").
		Set("user-data-dir", "path").
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
		Set("start-maximized").
		Set("window-size", "1920,1080").
		Set("disable-blink-features", "AutomationControlled").
		Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36").
		MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect().NoDefaultDevice()
	page := browser.MustPage("https://meet.google.com/nxf-prkz-bnw")

	//page.MustElement("[name='login']").MustClick()
	err := page.WaitLoad()
	if err != nil {
		log.Println(err.Error())
		return
	}

	page.MustWaitLoad().MustScreenshot("a.png")
	time.Sleep(time.Hour)
}
