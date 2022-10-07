package main

import (
	"errors"
	"github.com/dunkbing/meeting-bot/pkg/config"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dunkbing/meeting-bot/pkg/recorder"
)

func getConfig(configFile, configBody string) (*config.Config, error) {
	if configBody == "" {
		if configFile != "" {
			content, err := os.ReadFile(configFile)
			if err != nil {
				return nil, err
			}
			configBody = string(content)
		} else {
			return nil, errors.New("missing config")
		}
	}

	config.SetConfigBody(configBody)
	return config.GetConfig()
}

func runRecorder(c *cli.Context) error {
	configFile := c.String("config")
	configBody := c.String("config-body")
	_, err := getConfig(configFile, configBody)
	if err != nil {
		return err
	}

	rec := recorder.NewRecorder("standalone")

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-stopChan
		logrus.Info("exit requested, stopping recording and shutting down", "signal", sig)
		rec.Stop()
	}()

	res := rec.Run()
	if res.Error == "" {
		return nil
	}
	return errors.New(res.Error)
}
