package main

import (
	"fmt"
	"os"

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
				EnvVars: []string{"CONFIG_BODY"},
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
		Action:  runRecorder,
		Version: "1",
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
