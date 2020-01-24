package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include "gst.h"

*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/pion/webrtc/v2"
	"github.com/pkg/errors"

	"n.eko.moe/neko/internal/types"
)

/*
	apt-get install \
		libgstreamer1.0-0 \
		gstreamer1.0-plugins-base \
		gstreamer1.0-plugins-good \
		gstreamer1.0-plugins-bad \
		gstreamer1.0-plugins-ugly\
		gstreamer1.0-libav \
		gstreamer1.0-doc \
		gstreamer1.0-tools \
		gstreamer1.0-x \
		gstreamer1.0-alsa \
		gstreamer1.0-pulseaudio
*/

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	Pipeline  *C.GstElement
	Sample    chan types.Sample
	CodecName string
	ClockRate float32
	id        int
}

var pipelines = make(map[int]*Pipeline)
var pipelinesLock sync.Mutex
var registry *C.GstRegistry

const (
	videoClockRate = 90000
	audioClockRate = 48000
	pcmClockRate   = 8000
)

func init() {
	C.gstreamer_init()
	registry = C.gst_registry_get()
}

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(codecName string, pipelineSrc string) (*Pipeline, error) {
	pipelineStr := "appsink name=appsink"
	var clockRate float32

	switch codecName {
	case webrtc.VP8:
		// https://gstreamer.freedesktop.org/documentation/vpx/vp8enc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1
		pipelineStr = pipelineSrc + " ! vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1 ! " + pipelineStr
		clockRate = videoClockRate

		if err := CheckPlugins([]string{"ximagesrc", "vpx"}); err != nil {
			return nil, err
		}

	case webrtc.VP9:
		// https://gstreamer.freedesktop.org/documentation/vpx/vp9enc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// vp9enc

		// Causes panic!
		pipelineStr = pipelineSrc + " ! vp9enc ! " + pipelineStr
		clockRate = videoClockRate

		if err := CheckPlugins([]string{"ximagesrc", "vpx"}); err != nil {
			return nil, err
		}

	case webrtc.H264:
		// https://gstreamer.freedesktop.org/documentation/x264/index.html?gi-language=c
		// gstreamer1.0-plugins-ugly
		// video/x-raw,format=I420 ! x264enc bframes=0 key-int-max=60 byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream

		// https://gstreamer.freedesktop.org/documentation/openh264/openh264enc.html?gi-language=c#openh264enc
		// gstreamer1.0-plugins-bad
		// openh264enc multi-thread=4 complexity=high bitrate=3072000 max-bitrate=4096000
		pipelineStr = pipelineSrc + " ! openh264enc multi-thread=4 complexity=high bitrate=3072000 max-bitrate=4096000 ! video/x-h264,stream-format=byte-stream ! " + pipelineStr
		clockRate = videoClockRate

		if err := CheckPlugins([]string{"ximagesrc"}); err != nil {
			return nil, err
		}

		if err := CheckPlugins([]string{"openh264"}); err != nil {
			pipelineStr = pipelineSrc + " ! video/x-raw,format=I420 ! x264enc bframes=0 key-int-max=60 byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream ! " + pipelineStr

			if err := CheckPlugins([]string{"x264"}); err != nil {
				return nil, err
			}
		}

	case webrtc.Opus:
		// https://gstreamer.freedesktop.org/documentation/opus/opusenc.html
		// gstreamer1.0-plugins-base
		// opusenc
		pipelineStr = pipelineSrc + " ! opusenc ! " + pipelineStr
		clockRate = audioClockRate

		if err := CheckPlugins([]string{"pulseaudio", "opus"}); err != nil {
			return nil, err
		}

	case webrtc.G722:
		// https://gstreamer.freedesktop.org/documentation/libav/avenc_g722.html?gi-language=c
		// gstreamer1.0-libav
		// avenc_g722
		pipelineStr = pipelineSrc + " ! avenc_g722 ! " + pipelineStr
		clockRate = audioClockRate

		if err := CheckPlugins([]string{"pulseaudio", "libav"}); err != nil {
			return nil, err
		}

	case webrtc.PCMU:
		// https://gstreamer.freedesktop.org/documentation/mulaw/mulawenc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// audio/x-raw, rate=8000 ! mulawenc

		pipelineStr = pipelineSrc + " ! audio/x-raw, rate=8000 ! mulawenc ! " + pipelineStr
		clockRate = pcmClockRate

		if err := CheckPlugins([]string{"pulseaudio", "mulaw"}); err != nil {
			return nil, err
		}

	case webrtc.PCMA:
		// https://gstreamer.freedesktop.org/documentation/alaw/alawenc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// audio/x-raw, rate=8000 ! alawenc
		pipelineStr = pipelineSrc + " ! audio/x-raw, rate=8000 ! alawenc ! " + pipelineStr
		clockRate = pcmClockRate

		if err := CheckPlugins([]string{"pulseaudio", "alaw"}); err != nil {
			return nil, err
		}

	default:
		return nil, errors.Errorf("unknown video codec %s", codecName)
	}

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	pipeline := &Pipeline{
		Pipeline:  C.gstreamer_send_create_pipeline(pipelineStrUnsafe),
		Sample:    make(chan types.Sample),
		CodecName: codecName,
		ClockRate: clockRate,
		id:        len(pipelines),
	}

	pipelines[pipeline.id] = pipeline
	return pipeline, nil
}

// Start starts the GStreamer Pipeline
func (p *Pipeline) Start() {
	C.gstreamer_send_start_pipeline(p.Pipeline, C.int(p.id))
}

// Stop stops the GStreamer Pipeline
func (p *Pipeline) Stop() {
	C.gstreamer_send_stop_pipeline(p.Pipeline)
}

func CheckPlugins(plugins []string) error {
	var plugin *C.GstPlugin
	for _, pluginstr := range plugins {
		plugincstr := C.CString(pluginstr)
		plugin = C.gst_registry_find_plugin(registry, plugincstr)
		C.free(unsafe.Pointer(plugincstr))
		if plugin == nil {
			return fmt.Errorf("required gstreamer plugin %s not found", pluginstr)
		}
	}

	return nil
}

//export goHandlePipelineBuffer
func goHandlePipelineBuffer(buffer unsafe.Pointer, bufferLen C.int, duration C.int, pipelineID C.int) {
	pipelinesLock.Lock()
	pipeline, ok := pipelines[int(pipelineID)]
	pipelinesLock.Unlock()

	if ok {
		samples := uint32(pipeline.ClockRate * (float32(duration) / 1000000000))
		pipeline.Sample <- types.Sample{Data: C.GoBytes(buffer, bufferLen), Samples: samples}
	} else {
		fmt.Printf("discarding buffer, no pipeline with id %d", int(pipelineID))
	}
	C.free(buffer)
}
