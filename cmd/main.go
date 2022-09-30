package main

import (
	"github.com/dunkbing/meeting-bot/pkg/config"
	"github.com/dunkbing/meeting-bot/pkg/recorder"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logrus.SetReportCaller(true)
	logrus.Infoln("start")
	_, _ = config.Get()
	//b := bot.New()
	//if b == nil {
	//	return
	//}
	//if b.Type == bot.GoogleMeet {
	//	b.JoinGoogleMeet()
	//}

	rec := recorder.NewRecorder("vip")

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-stopChan
		logrus.Infoln("exit requested, stopping recording and shutting down", "signal", sig)
		rec.Stop()
	}()

	err := rec.Run()
	if err != nil {
		logrus.Errorln("app error:", err.Error())
	}
}
