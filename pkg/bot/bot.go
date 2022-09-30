package bot

import (
	"github.com/dunkbing/meeting-bot/pkg/config"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"strings"
	"time"
)

type meetingType string

const (
	GoogleMeet  meetingType = "google_meet"
	Teams       meetingType = "teams"
	Zoom        meetingType = "zoom"
	InvalidType meetingType = ""
)

type bot struct {
	browser         *rod.Browser
	meetingUrl      string
	meetingStarted  bool
	meetingFinished bool
	Type            meetingType
}

func (b *bot) WaitForApproval() {

}

func (b *bot) JoinGoogleMeet() {
	cfg, _ := config.Get()
	page := b.browser.MustPage(b.meetingUrl)
	page.MustElementX("//input[@placeholder='Your name']").MustInput(cfg.MeetingUsername)
	page.MustElement(".jtn8y")
	page.MustElementX("//span[contains(text(), \"Ask to join\")]/parent::button").MustClick()
	time.Sleep(time.Hour)
}

func GetMeetingType(url string) meetingType {
	if strings.Contains(url, "zoom.us") {
		return Zoom
	}
	switch {
	case strings.Contains(url, "zoom.us"):
		return Zoom
	case strings.Contains(url, "google"):
		return GoogleMeet
	case strings.Contains(url, "microsoft") || strings.Contains(url, "teams"):
		return Teams
	default:
		return InvalidType
	}
}

func New() *bot {
	cfg, _ := config.Get()
	type_ := GetMeetingType(cfg.MeetingUrl)

	if type_ == InvalidType {
		return nil
	}

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
		Set("disable-blink-features", "AutomationControlled").
		Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36").
		MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect().NoDefaultDevice()
	return &bot{
		browser:         browser,
		meetingUrl:      cfg.MeetingUrl,
		meetingStarted:  false,
		meetingFinished: false,
		Type:            type_,
	}
}
