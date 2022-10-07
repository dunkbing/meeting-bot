//go:build !test

package gst

import (
	"fmt"
	"github.com/dunkbing/meeting-bot/pkg/config"

	"github.com/tinyzimmer/go-gst/gst"
)

type InputBin struct {
	isStream      bool
	bin           *gst.Bin
	audioElements []*gst.Element
	videoElements []*gst.Element
	audioQueue    *gst.Element
	videoQueue    *gst.Element
	mux           *gst.Element
}

func newVideoElements() ([]*gst.Element, *gst.Element, error) {
	conf, _ := config.GetConfig()
	// create video elements
	xImageSrc, err := gst.NewElement("ximagesrc")
	if err != nil {
		return nil, nil, err
	}
	err = xImageSrc.SetProperty("use-damage", false)
	if err != nil {
		return nil, nil, err
	}
	err = xImageSrc.SetProperty("show-pointer", false)
	if err != nil {
		return nil, nil, err
	}

	videoConvert, err := gst.NewElement("videoconvert")
	if err != nil {
		return nil, nil, err
	}

	framerateCaps, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, nil, err
	}
	err = framerateCaps.SetProperty("caps", gst.NewCapsFromString(
		fmt.Sprintf("video/x-raw,framerate=%d/1", conf.Media.Framerate),
	))
	if err != nil {
		return nil, nil, err
	}

	x264Enc, err := gst.NewElement("x264enc")
	if err != nil {
		return nil, nil, err
	}
	if err = x264Enc.SetProperty("bitrate", uint(conf.Media.VideoBitrate)); err != nil {
		return nil, nil, err
	}
	x264Enc.SetArg("speed-preset", "veryfast")
	x264Enc.SetArg("tune", "zerolatency")

	profileCaps, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, nil, err
	}
	err = profileCaps.SetProperty("caps", gst.NewCapsFromString(
		fmt.Sprintf(
			"video/x-h264,profile=%s,framerate=%d/1",
			conf.Media.Profile,
			conf.Media.Framerate,
		),
	))
	if err != nil {
		return nil, nil, err
	}

	videoQueue, err := gst.NewElement("queue")
	if err != nil {
		return nil, nil, err
	}
	if err = videoQueue.SetProperty("max-size-time", uint64(3e9)); err != nil {
		return nil, nil, err
	}
	return []*gst.Element{xImageSrc, videoConvert, framerateCaps, x264Enc, profileCaps, videoQueue}, videoQueue, nil
}

func newAudioElements() ([]*gst.Element, *gst.Element, error) {
	conf, err := config.GetConfig()
	// create audio elements
	pulseSrc, err := gst.NewElement("pulsesrc")
	if err != nil {
		return nil, nil, err
	}

	audioConvert, err := gst.NewElement("audioconvert")
	if err != nil {
		return nil, nil, err
	}

	audioCapsFilter, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, nil, err
	}
	err = audioCapsFilter.SetProperty("caps", gst.NewCapsFromString(
		fmt.Sprintf("audio/x-raw,format=S16LE,layout=interleaved,rate=%d,channels=2", conf.Media.AudioFrequency),
	))
	if err != nil {
		return nil, nil, err
	}

	faac, err := gst.NewElement("faac")
	if err != nil {
		return nil, nil, err
	}
	err = faac.SetProperty("bitrate", int(conf.Media.AudioBitrate*1000))
	if err != nil {
		return nil, nil, err
	}

	audioQueue, err := gst.NewElement("queue")
	if err != nil {
		return nil, nil, err
	}
	if err = audioQueue.SetProperty("max-size-time", uint64(3e9)); err != nil {
		return nil, nil, err
	}
	return []*gst.Element{pulseSrc, audioConvert, audioCapsFilter, faac, audioQueue}, audioQueue, nil
}

func newMuxElement(isStream bool) (*gst.Element, error) {
	// create mux
	var mux *gst.Element
	var err error
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
	return mux, nil
}

func newVideoInputBin(isStream bool) (*InputBin, error) {
	audioElements, audioQueue, err := newAudioElements()
	videoElements, videoQueue, err := newVideoElements()
	mux, err := newMuxElement(isStream)

	elements := append(audioElements, videoElements...)
	elements = append(elements, mux)

	// create bin
	bin := gst.NewBin("input")
	err = bin.AddMany(elements...)
	if err != nil {
		return nil, err
	}

	// create ghost pad
	ghostPad := gst.NewGhostPad("src", mux.GetStaticPad("src"))
	if !bin.AddPad(ghostPad.Pad) {
		return nil, ErrGhostPadFailed
	}

	return &InputBin{
		isStream:      isStream,
		bin:           bin,
		audioElements: audioElements,
		videoElements: videoElements,
		audioQueue:    audioQueue,
		videoQueue:    videoQueue,
		mux:           mux,
	}, nil
}

func newAudioInputBin(isStream bool) (*InputBin, error) {
	audioElements, audioQueue, err := newAudioElements()
	if err != nil {
		return nil, err
	}
	mux, err := newMuxElement(isStream)

	audioElements = append(audioElements, mux)

	// create bin
	bin := gst.NewBin("input")
	err = bin.AddMany(audioElements...)
	if err != nil {
		return nil, err
	}

	// create ghost pad
	ghostPad := gst.NewGhostPad("src", mux.GetStaticPad("src"))
	if !bin.AddPad(ghostPad.Pad) {
		return nil, ErrGhostPadFailed
	}

	return &InputBin{
		isStream:      isStream,
		bin:           bin,
		audioElements: audioElements,
		audioQueue:    audioQueue,
		mux:           mux,
	}, nil
}

func (b *InputBin) Link() error {
	// link audio elements
	if err := gst.ElementLinkMany(b.audioElements...); err != nil {
		return err
	}

	// link video elements
	if b.videoElements != nil {
		if err := gst.ElementLinkMany(b.videoElements...); err != nil {
			return err
		}
	}

	// link audio and video queues to mux
	var muxAudioPad, muxVideoPad *gst.Pad
	if b.isStream {
		muxAudioPad = b.mux.GetRequestPad("audio")
		muxVideoPad = b.mux.GetRequestPad("video")
	} else {
		muxAudioPad = b.mux.GetRequestPad("audio_%u")
		muxVideoPad = b.mux.GetRequestPad("video_%u")
	}
	if err := requireLink(b.audioQueue.GetStaticPad("src"), muxAudioPad); err != nil {
		fmt.Println("link audio err", err)
		return err
	}
	if b.videoQueue != nil {
		if err := requireLink(b.videoQueue.GetStaticPad("src"), muxVideoPad); err != nil {
			fmt.Println("link video err")
			return err
		}
	}

	return nil
}
