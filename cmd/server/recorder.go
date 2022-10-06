package main

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dunkbing/meeting-bot/pkg/recorder"
)

func runRecorder(c *cli.Context) error {
	conf, err := getConfig(c)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	rec := recorder.NewRecorder(conf, "standalone")

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-stopChan
		logrus.Info("exit requested, stopping recording and shutting down", "signal", sig)
		rec.Stop()
	}()

	res := rec.Run()
	//service.LogResult(res)
	if res.Error == "" {
		return nil
	}
	return errors.New(res.Error)
}
