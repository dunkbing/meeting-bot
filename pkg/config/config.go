package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	PageLoadTimeout     int
	WaitApprovalTimeout int
	MeetingUrl          string
	MeetingUsername     string
	MaxMeetingDuration  int
	MaxAloneTime        int
	MaxSilentDuration   int
	Display             string
}

func (c *Config) initDisplay() error {
	d := os.Getenv("DISPLAY")
	if d != "" && strings.HasPrefix(d, ":") {
		num, err := strconv.Atoi(d[1:])
		if err == nil && num > 0 && num <= 2147483647 {
			c.Display = d
			return nil
		}
	}

	if c.Display == "" {
		rand.Seed(time.Now().UnixNano())
		display := fmt.Sprintf(":%d", 10+rand.Intn(2147483637))
		fmt.Println("Display:", display)
		c.Display = display
	}

	// GStreamer uses display from env
	if err := os.Setenv("DISPLAY", c.Display); err != nil {
		return err
	}

	return nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

var _cfg *Config

// Get returns app config.
func Get() (*Config, error) {
	if _cfg != nil {
		return _cfg, nil
	}
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %s\n", err.Error())
	}
	pageLoadTimeout, _ := strconv.Atoi(getEnv("PAGE_LOAD_TIMEOUT", "90"))
	waitApprovalTimeout, _ := strconv.Atoi(getEnv("WAIT_APPROVAL_TIMEOUT", "180"))
	meetingUrl := getEnv("MEETING_URL", "")
	meetingUsername := getEnv("MEETING_USERNAME", "")
	meetingDuration, _ := strconv.Atoi(getEnv("MAX_MEETING_DURATION", "7200"))
	maxAloneTime, _ := strconv.Atoi(getEnv("MAX_ALONE_TIME", "120"))
	maxSilentDuration, _ := strconv.Atoi(getEnv("MAX_SILENT_DURATION", "300"))
	cfg := &Config{
		PageLoadTimeout:     pageLoadTimeout,
		WaitApprovalTimeout: waitApprovalTimeout,
		MeetingUrl:          meetingUrl,
		MeetingUsername:     meetingUsername,
		MaxMeetingDuration:  meetingDuration,
		MaxAloneTime:        maxAloneTime,
		MaxSilentDuration:   maxSilentDuration,
	}
	_cfg = cfg
	err = _cfg.initDisplay()
	if err != nil {
		logrus.Errorln("Error initDisplay:", err)
		return nil, err
	}

	return cfg, nil
}
