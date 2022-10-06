package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/dunkbing/meeting-bot/pkg/config"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "meeting-bot",
		Usage:       "Meeting Bot",
		Description: "runs bot in standalone mode",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "path to LiveKit recording config defaults",
			},
			&cli.StringFlag{
				Name:    "config-body",
				Usage:   "Default LiveKit recording config in JSON, typically passed in as an env var in a container",
				EnvVars: []string{"LIVEKIT_RECORDER_CONFIG"},
			},
			&cli.StringFlag{
				Name:  "request",
				Usage: "path to json StartRecordingRequest file",
			},
			&cli.StringFlag{
				Name:    "request-body",
				Usage:   "StartRecordingRequest json",
				EnvVars: []string{"RECORDING_REQUEST"},
			},
		},
		Action:  run,
		Version: "1",
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func run(c *cli.Context) error {
	return runRecorder(c)
}

func getConfig(c *cli.Context) (*config.Config, error) {
	configFile := c.String("config")
	configBody := c.String("config-body")
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

	return config.NewConfig(configBody)
}
