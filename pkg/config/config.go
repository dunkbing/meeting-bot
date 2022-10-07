package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	ProfileBaseline = "baseline"
	ProfileMain     = "main"
	ProfileHigh     = "high"
)

var validProfiles = map[string]bool{
	ProfileBaseline: true,
	ProfileMain:     true,
	ProfileHigh:     true,
}

type Config struct {
	LogLevel   string           `yaml:"log_level"`
	FileOutput FileOutputConfig `yaml:"file_output"`
	Screen     ScreenConfig     `yaml:"screen"`
	Media      MediaConfig      `yaml:"media"`
	Bot        BotConfig        `yaml:"bot"`
}

type FileOutputConfig struct {
	Local     bool          `yaml:"local"`
	S3        *S3Config     `yaml:"s3"`
	AzBlob    *AzBlobConfig `yaml:"azblob"`
	GCPConfig *GCPConfig    `yaml:"gcp"`
	FileDir   string        `yaml:"file_dir"`
	FileName  string        `yaml:"file_name"`
}

type S3Config struct {
	AccessKey string `yaml:"access_key"`
	Secret    string `yaml:"secret"`
	Endpoint  string `yaml:"endpoint"`
	Region    string `yaml:"region"`
	Bucket    string `yaml:"bucket"`
}

type AzBlobConfig struct {
	AccountName   string `yaml:"account_name"`
	AccountKey    string `yaml:"account_key"`
	ContainerName string `yaml:"container_name"`
}

type GCPConfig struct {
	Bucket string `yaml:"bucket"`
}

type MediaConfig struct {
	Framerate      int32  `yaml:"framerate"`
	AudioBitrate   int32  `yaml:"audio_bitrate"`
	AudioFrequency int32  `yaml:"audio_frequency"`
	VideoBitrate   int32  `yaml:"video_bitrate"`
	Profile        string `yaml:"profile"`
}

type ScreenConfig struct {
	Width    int32  `yaml:"width"`
	Height   int32  `yaml:"height"`
	Depth    int32  `yaml:"depth"`
	Display  string `yaml:"-"`
	Insecure bool   `yaml:"insecure"`
}

type BotConfig struct {
	MeetingUrl string `yaml:"meeting_url"`
	Username   string `yaml:"username"`
	Type       string `yaml:"type"`
}

func (c *Config) initDisplay() error {
	d := os.Getenv("DISPLAY")
	if d != "" && strings.HasPrefix(d, ":") {
		num, err := strconv.Atoi(d[1:])
		if err == nil && num > 0 && num <= 2147483647 {
			c.Screen.Display = d
			return nil
		}
	}

	if c.Screen.Display == "" {
		rand.Seed(time.Now().UnixNano())
		c.Screen.Display = fmt.Sprintf(":%d", 10+rand.Intn(2147483637))
	}

	// GStreamer uses display from env
	if err := os.Setenv("DISPLAY", c.Screen.Display); err != nil {
		return err
	}

	return nil
}

var _config *Config
var _configBody string

func SetConfigBody(_conf string) {
	_configBody = _conf
}

func GetConfig() (*Config, error) {
	if _config != nil {
		return _config, nil
	}
	// start with defaults
	conf := &Config{
		LogLevel: "info",
		Screen: ScreenConfig{
			Width:  1920,
			Height: 1080,
			Depth:  24,
		},
		Media: MediaConfig{
			Framerate:      30,
			AudioBitrate:   128,
			AudioFrequency: 44100,
			VideoBitrate:   4500,
			Profile:        ProfileMain,
		},
	}

	if _configBody != "" {
		if err := yaml.Unmarshal([]byte(_configBody), conf); err != nil {
			return nil, fmt.Errorf("could not parse config: %v", err)
		}
	}

	// apply preset options
	if conf.FileOutput.S3 == nil && conf.FileOutput.AzBlob == nil && conf.FileOutput.GCPConfig == nil {
		conf.FileOutput.Local = true
	}

	if !validProfiles[conf.Media.Profile] {
		return nil, fmt.Errorf("invalid profile %s", conf.Media.Profile)
	}

	// GStreamer log level
	if os.Getenv("GST_DEBUG") == "" {
		var gstDebug int
		switch conf.LogLevel {
		case "debug":
			gstDebug = 2
			logrus.SetLevel(logrus.DebugLevel)
		case "info", "warn", "error":
			gstDebug = 1
			logrus.SetLevel(logrus.InfoLevel)
		case "panic":
			gstDebug = 0
			logrus.SetLevel(logrus.PanicLevel)
		}
		if err := os.Setenv("GST_DEBUG", fmt.Sprint(gstDebug)); err != nil {
			return nil, err
		}
	}

	err := conf.initDisplay()
	_config = conf
	return conf, err
}

func TestConfig() (*Config, error) {
	conf := &Config{
		LogLevel: "debug",
		Screen: ScreenConfig{
			Width:  1920,
			Height: 1080,
			Depth:  24,
		},
		Media: MediaConfig{
			Framerate:      30,
			AudioBitrate:   128,
			AudioFrequency: 44100,
			VideoBitrate:   4500,
			Profile:        ProfileMain,
		},
	}
	err := conf.initDisplay()
	return conf, err
}
