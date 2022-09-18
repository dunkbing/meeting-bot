package main

import (
	"github.com/dunkbing/meeting-bot/bot"
	"github.com/dunkbing/meeting-bot/config"
)

func main() {
	_, _ = config.Get()
	b := bot.New()
	if b == nil {
		return
	}
	if b.Type == bot.GoogleMeet {
		b.JoinGoogleMeet()
	}
}
