package config_test

import (
	"testing"

	"github.com/dunkbing/meeting-bot/pkg/config"
	"github.com/stretchr/testify/require"
)

var testConfig = `
log_level: debug
api_key: key
api_secret: secret
ws_url: wss://localhost:7880
file_output:
  local: true
redis:
  address: 192.168.65.2:6379
defaults:
  width: 320
  height: 200
  depth: 24
  framerate: 10
  audio_bitrate: 96
  audio_frequency: 22050
  video_bitrate: 750
  profile: high
`

var testRequests = []string{`
{
	"template": {
		"layout": "speaker-dark",
		"room_name": "test-room"
	},
	"filepath": "/out/filename.mp4",
	"options": {
		"preset": "FULL_HD_30"
	}
}
`, `
{
	"url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
	"rtmp": {
        "urls": ["rtmp://stream-url.com", "rtmp://live.twitch.tv/app/stream-key"]
    },
	"options": {
		"audio_bitrate": 96,
		"audio_frequency": 22050,
		"video_bitrate": 750,
		"profile": "main"
	}
}
`}

func TestConfig(t *testing.T) {
	config.SetConfigBody(testConfig)
	conf, err := config.GetConfig()
	require.NoError(t, err)
	require.Equal(t, int32(320), conf.Screen.Width)
	require.Equal(t, int32(96), conf.Media.AudioBitrate)
	require.Equal(t, config.ProfileHigh, conf.Media.Profile)
}
