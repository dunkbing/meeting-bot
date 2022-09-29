package pipeline

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tinyzimmer/go-gst/gst"
)

type Options struct {
	Width          int
	Height         int
	Depth          int
	Framerate      int
	AudioBitrate   int
	AudioFrequency int
	VideoBitrate   int
	Profile        string
}

type InputBin struct {
	isStream      bool
	bin           *gst.Bin
	audioElements []*gst.Element
	videoElements []*gst.Element
	audioQueue    *gst.Element
	videoQueue    *gst.Element
	mux           *gst.Element
}

func newInputBin(isStream bool, options Options) (*InputBin, error) {
	// region audio elements
	pulseSrc, err := gst.NewElement("pulsesrc")
	if err != nil {
		return nil, err
	}

	audioConvert, err := gst.NewElement("audioconvert")
	if err != nil {
		return nil, err
	}

	audioCapsFilter, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, err
	}
	err = audioCapsFilter.SetProperty("caps", gst.NewCapsFromString(
		fmt.Sprintf("audio/x-raw,format=S16LE,layout=interleaved,rate=%d,channels=2", options.AudioFrequency),
	))
	if err != nil {
		return nil, err
	}

	faac, err := gst.NewElement("lamemp3enc")
	if err != nil {
		return nil, err
	}
	err = faac.SetProperty("bitrate", options.AudioBitrate)
	if err != nil {
		return nil, err
	}

	audioQueue, err := gst.NewElement("queue")
	if err != nil {
		return nil, err
	}
	if err = audioQueue.SetProperty("max-size-time", uint64(3e9)); err != nil {
		return nil, err
	}
	// endregion

	// region video elements
	xImageSrc, err := gst.NewElement("ximagesrc")
	if err != nil {
		return nil, err
	}
	err = xImageSrc.SetProperty("use-damage", true)
	if err != nil {
		return nil, err
	}
	err = xImageSrc.SetProperty("show-pointer", false)
	if err != nil {
		return nil, err
	}

	videoConvert, err := gst.NewElement("videoconvert")
	if err != nil {
		return nil, err
	}

	framerateCaps, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, err
	}
	err = framerateCaps.SetProperty("caps", gst.NewCapsFromString(
		fmt.Sprintf("video/x-raw,framerate=%d/1", options.Framerate),
	))
	if err != nil {
		return nil, err
	}

	x264Enc, err := gst.NewElement("x264enc")
	if err != nil {
		return nil, err
	}
	if err = x264Enc.SetProperty("bitrate", uint(options.VideoBitrate)); err != nil {
		return nil, err
	}
	x264Enc.SetArg("speed-preset", "veryfast")
	x264Enc.SetArg("tune", "zerolatency")

	profileCaps, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, err
	}
	err = profileCaps.SetProperty("caps", gst.NewCapsFromString(
		fmt.Sprintf("video/x-h264,profile=%s,framerate=%d/1", options.Profile, options.Framerate),
	))
	if err != nil {
		return nil, err
	}

	videoQueue, err := gst.NewElement("queue")
	if err != nil {
		return nil, err
	}
	if err = videoQueue.SetProperty("max-size-time", uint64(3e9)); err != nil {
		return nil, err
	}
	// endregion

	// create mux
	var mux *gst.Element
	if isStream {
		mux, err = gst.NewElement("flvmux")
		if err != nil {
			return nil, err
		}
		err = mux.Set("streamable", true)
		if err != nil {
			return nil, err
		}
	} else {
		mux, err = gst.NewElement("mp4mux")
		if err != nil {
			return nil, err
		}
		err = mux.SetProperty("faststart", true)
		if err != nil {
			return nil, err
		}
	}

	// create bin
	bin := gst.NewBin("input")
	err = bin.AddMany(
		// audio
		pulseSrc, audioConvert, audioCapsFilter, faac, audioQueue,
		// video
		xImageSrc, videoConvert, framerateCaps, x264Enc, profileCaps, videoQueue,
		// mux
		mux,
	)
	if err != nil {
		return nil, err
	}

	// create ghost pad
	ghostPad := gst.NewGhostPad("src", mux.GetStaticPad("src"))
	if !bin.AddPad(ghostPad.Pad) {
		return nil, errors.New("failed to add ghost pad to bin")
	}

	return &InputBin{
		isStream:      isStream,
		bin:           bin,
		audioElements: []*gst.Element{pulseSrc, audioConvert, audioCapsFilter, faac, audioQueue},
		videoElements: []*gst.Element{xImageSrc, videoConvert, framerateCaps, x264Enc, profileCaps, videoQueue},
		audioQueue:    audioQueue,
		videoQueue:    videoQueue,
		mux:           mux,
	}, nil
}

func (b *InputBin) Link() error {
	// link audio elements
	if err := gst.ElementLinkMany(b.audioElements...); err != nil {
		logrus.Errorln("Error link audio elements")
		return err
	}

	// link video elements
	if err := gst.ElementLinkMany(b.videoElements...); err != nil {
		logrus.Errorln("Error link video elements")
		return err
	}

	// link audio and video queues to mux
	var muxAudioPad *gst.Pad
	var muxVideoPad *gst.Pad
	if b.isStream {
		muxAudioPad = b.mux.GetRequestPad("audio")
		muxVideoPad = b.mux.GetRequestPad("video")
	} else {
		muxAudioPad = b.mux.GetRequestPad("audio_%u")
		muxVideoPad = b.mux.GetRequestPad("video_%u")
	}
	if err := requireLink(b.audioQueue.GetStaticPad("src"), muxAudioPad); err != nil {
		return err
	}
	if err := requireLink(b.videoQueue.GetStaticPad("src"), muxVideoPad); err != nil {
		return err
	}
	return nil
}
