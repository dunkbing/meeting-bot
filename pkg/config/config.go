package config

import (
	"fmt"
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
	LogLevel   string      `yaml:"log_level"`
	Insecure   bool        `yaml:"insecure"`
	Redis      RedisConfig `yaml:"redis"`
	FileOutput FileOutput  `yaml:"file_output"`
	Defaults   Defaults    `yaml:"defaults"`
	Display    string      `yaml:"-"`
}

type RedisConfig struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type FileOutput struct {
	Local     bool          `yaml:"local"`
	S3        *S3Config     `yaml:"s3"`
	AzBlob    *AzblobConfig `yaml:"azblob"`
	GCPConfig *GCPConfig    `yaml:"gcp"`
}

type S3Config struct {
	AccessKey string `yaml:"access_key"`
	Secret    string `yaml:"secret"`
	Endpoint  string `yaml:"endpoint"`
	Region    string `yaml:"region"`
	Bucket    string `yaml:"bucket"`
}

type AzblobConfig struct {
	AccountName   string `yaml:"account_name"`
	AccountKey    string `yaml:"account_key"`
	ContainerName string `yaml:"container_name"`
}

type GCPConfig struct {
	Bucket string `yaml:"bucket"`
}

type Defaults struct {
	Width          int32  `yaml:"width"`
	Height         int32  `yaml:"height"`
	Depth          int32  `yaml:"depth"`
	Framerate      int32  `yaml:"framerate"`
	AudioBitrate   int32  `yaml:"audio_bitrate"`
	AudioFrequency int32  `yaml:"audio_frequency"`
	VideoBitrate   int32  `yaml:"video_bitrate"`
	Profile        string `yaml:"profile"`
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
		c.Display = fmt.Sprintf(":%d", 10+rand.Intn(2147483637))
	}

	// GStreamer uses display from env
	if err := os.Setenv("DISPLAY", c.Display); err != nil {
		return err
	}

	return nil
}

func NewConfig(confString string) (*Config, error) {
	// start with defaults
	conf := &Config{
		LogLevel: "info",
		Defaults: Defaults{
			Width:          1920,
			Height:         1080,
			Depth:          24,
			Framerate:      30,
			AudioBitrate:   128,
			AudioFrequency: 44100,
			VideoBitrate:   4500,
			Profile:        ProfileMain,
		},
	}

	if confString != "" {
		if err := yaml.Unmarshal([]byte(confString), conf); err != nil {
			return nil, fmt.Errorf("could not parse config: %v", err)
		}
	}

	// apply preset options
	if conf.FileOutput.S3 == nil && conf.FileOutput.AzBlob == nil && conf.FileOutput.GCPConfig == nil {
		conf.FileOutput.Local = true
	}

	if !validProfiles[conf.Defaults.Profile] {
		return nil, fmt.Errorf("invalid profile %s", conf.Defaults.Profile)
	}

	// GStreamer log level
	if os.Getenv("GST_DEBUG") == "" {
		var gstDebug int
		switch conf.LogLevel {
		case "debug":
			gstDebug = 2
		case "info", "warn", "error":
			gstDebug = 1
		case "panic":
			gstDebug = 0
		}
		if err := os.Setenv("GST_DEBUG", fmt.Sprint(gstDebug)); err != nil {
			return nil, err
		}
	}

	err := conf.initDisplay()
	return conf, err
}

func TestConfig() (*Config, error) {
	conf := &Config{
		LogLevel: "debug",
		Redis: RedisConfig{
			Address: "localhost:6379",
		},
		Defaults: Defaults{
			Width:          1920,
			Height:         1080,
			Depth:          24,
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
