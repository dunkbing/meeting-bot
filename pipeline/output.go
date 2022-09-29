package pipeline

import (
	"errors"
	"github.com/tinyzimmer/go-gst/gst"
)

type OutputBin struct {
	isStream bool
	bin      *gst.Bin
	fileSink *gst.Element
}

func newFileOutputBin(filename string) (*OutputBin, error) {
	// create elements
	sink, err := gst.NewElement("filesink")
	if err != nil {
		return nil, err
	}
	if err = sink.SetProperty("location", filename); err != nil {
		return nil, err
	}
	if err = sink.SetProperty("sync", false); err != nil {
		return nil, err
	}

	// create bin
	bin := gst.NewBin("output")
	if err = bin.Add(sink); err != nil {
		return nil, err
	}

	// add ghost pad
	ghostPad := gst.NewGhostPad("sink", sink.GetStaticPad("sink"))
	if !bin.AddPad(ghostPad.Pad) {
		return nil, errors.New("failed to add ghost pad to bin")
	}

	return &OutputBin{
		isStream: false,
		bin:      bin,
		fileSink: sink,
	}, nil
}
