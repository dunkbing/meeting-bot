package bot

import (
	"errors"
	"fmt"
	"github.com/dunkbing/meeting-bot/pkg/config"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"strings"
	"time"
)

type MeetingType string

const (
	GoogleMeet     MeetingType = "google_meet"
	Teams          MeetingType = "teams"
	Zoom           MeetingType = "zoom"
	InvalidMeeting MeetingType = "invalid"
)

type Bot struct {
	browser         *rod.Browser
	meetingUrl      string
	meetingUsername string
	meetingType     MeetingType
	meetingStarted  bool
	meetingFinished bool
}

func (b *Bot) WaitForApproval(page *rod.Page) {
	var i int
	timeOut := time.Duration(30)
	var ele *rod.Element
	for start := time.Now(); time.Since(start) < time.Second*timeOut; {
		switch b.meetingType {
		case GoogleMeet:
			meetingDetailEle := ".r6xAKc"
			ele = page.MustElement(meetingDetailEle)
		case Teams:
			rosterBtn := "#roster-button"
			ele = page.MustElement(rosterBtn)
		case Zoom:
			numberCounter := ".footer-button__number-counter"
			ele = page.MustElement(numberCounter)
		}
		if ele != nil {
			break
		}
		time.Sleep(time.Second)
		i++
	}
}

func (b *Bot) JoinGoogleMeet() error {
	page := b.browser.MustPage(b.meetingUrl)
	inputUsernameXpath := "//input[@placeholder='Your name']"
	page.MustElementX(inputUsernameXpath).MustInput(b.meetingUsername)
	joinOptions := ".jtn8y"
	page.MustElement(joinOptions)

	currentUrl := page.MustInfo().URL
	if strings.Contains(currentUrl, "accounts") {
		return errors.New("cannot join organization meeting")
	}

	askToJoinBtn := "//span[contains(text(), \"Ask to join\")]/parent::button"
	page.MustElementX(askToJoinBtn).MustClick()

	b.WaitForApproval(page)

	return nil
}

func (b *Bot) JoinTeams() error {
	team := "https://teams.microsoft.com"
	live := "https://teams.live.com"
	if strings.Contains(b.meetingUrl, "teams.live") {
		b.meetingUrl = strings.Replace(b.meetingUrl, live, "", 1)
		b.meetingUrl = fmt.Sprintf("%s/_#%s?anon=true", live, b.meetingUrl)
	} else {
		b.meetingUrl = strings.Replace(b.meetingUrl, team, "", 1)
		b.meetingUrl = fmt.Sprintf("%s/_#%s&anon=true", team, b.meetingUrl)
	}
	fmt.Println(b.meetingUrl)
	page := b.browser.MustPage(b.meetingUrl)
	inputUsername := "#username"
	inputUsernameEle := page.MustElement(inputUsername)

	if strings.Contains(b.meetingUrl, "teams.live") {
		// TODO: press enter to dismiss popup
	}
	inputUsernameEle.MustInput(b.meetingUsername)

	joinBtn := ".join-btn"
	page.MustElement(joinBtn).MustClick()

	b.WaitForApproval(page)

	return nil
}

func (b *Bot) JoinZoom() error {
	page := b.browser.MustPage(b.meetingUrl)
	fmt.Println(page.String())
	return nil
}

func GetMeetingType(url string) MeetingType {
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
		return InvalidMeeting
	}
}

func (b *Bot) Start() error {
	switch b.meetingType {
	case GoogleMeet:
		return b.JoinGoogleMeet()
	case Teams:
		return b.JoinTeams()
	case Zoom:
		return b.JoinZoom()
	default:
		return nil
	}
}

func New() *Bot {
	cfg, _ := config.GetConfig()

	u := launcher.New().
		Bin("/usr/bin/google-chrome").
		Set("user-data-dir", "chrome-data").
		Set("no-first-run").
		Set("no-default-browser-check").
		Set("disable-gpu").
		Set("no-sandbox").
		Delete("--headless").
		Set("disable-infobars").
		Set("excludeSwitches", "enable-automation").
		Set("disable-background-networking").
		Set("enable-features", "NetworkService,NetworkServiceInProcess").
		Set("disable-background-timer-throttling").
		Set("disable-backgrounding-occluded-windows").
		Set("disable-breakpad").
		Set("disable-client-side-phishing-detection").
		Set("disable-default-apps").
		Set("disable-dev-shm-usage").
		Set("disable-extensions").
		Set("disable-features", "site-per-process,TranslateUI,BlinkGenPropertyTrees").
		Set("disable-hang-monitor").
		Set("disable-ipc-flooding-protection").
		Set("disable-popup-blocking").
		Set("disable-prompt-on-repost").
		Set("disable-renderer-backgrounding").
		Set("disable-sync").
		Set("force-color-profile", "srgb").
		Set("metrics-recording-only").
		Set("safebrowsing-disable-auto-update").
		Set("password-store", "basic").
		Set("use-mock-keychain").

		// custom args
		Set("kiosk").
		Set("autoplay-policy", "no-user-gesture-required").
		Set("window-position", "0,0").
		Set("window-size", fmt.Sprintf("%d,%d", cfg.Screen.Width, cfg.Screen.Height)).
		Set("display", cfg.Screen.Display).
		Set("incognito").
		Set("use-fake-ui-for-media-stream").
		Set("disable-software-rasterizer").
		Set("bwsi").
		Set("disable-blink-features", "AutomationControlled").
		Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36").
		MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect().NoDefaultDevice()
	return &Bot{
		browser:         browser,
		meetingUrl:      cfg.Bot.MeetingUrl,
		meetingUsername: cfg.Bot.Username,
		meetingType:     GetMeetingType(cfg.Bot.MeetingUrl),
		meetingStarted:  false,
		meetingFinished: false,
	}
}
