package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type Config struct {
	PageLoadTimeout     int
	WaitApprovalTimeout int
	MeetingUrl          string
	MeetingUsername     string
	MaxMeetingDuration  int
	MaxAloneTime        int
	MaxSilentDuration   int
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

	return cfg, nil
}
