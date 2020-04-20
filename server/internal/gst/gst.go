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

    gst-inspect-1.0 --version
    gst-inspect-1.0 plugin
    gst-launch-1.0 ximagesrc show-pointer=true use-damage=false ! video/x-raw,framerate=30/1 ! videoconvert ! queue ! vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1 ! autovideosink
    gst-launch-1.0 pulsesrc ! audioconvert ! opusenc ! autoaudiosink
*/

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	Pipeline  *C.GstElement
	Sample    chan types.Sample
	CodecName string
	ClockRate float32
	Src       string
	id        int
}

var pipelines = make(map[int]*Pipeline)
var pipelinesLock sync.Mutex
var registry *C.GstRegistry

const (
	videoClockRate = 90000
	audioClockRate = 48000
	pcmClockRate   = 8000
	videoSrc       = "ximagesrc xid=%s show-pointer=true use-damage=false ! video/x-raw ! videoconvert ! queue ! "
	audioSrc       = "pulsesrc device=%s ! audioconvert ! "
)

func init() {
	C.gstreamer_init()
	registry = C.gst_registry_get()
}

// CreateRTMPPipeline creates a GStreamer Pipeline
func CreateRTMPPipeline(pipelineDevice string, pipelineDisplay string, pipelineRTMP string) (*Pipeline, error) {
	video := fmt.Sprintf(videoSrc, pipelineDisplay)
	audio := fmt.Sprintf(audioSrc, pipelineDevice)
	return CreatePipeline(fmt.Sprintf("%s ! x264enc ! flv. ! %s ! faac ! flv. ! flvmux name='flv' ! rtmpsink location='%s'", video, audio, pipelineRTMP), "", 0)
}

// CreateAppPipeline creates a GStreamer Pipeline
func CreateAppPipeline(codecName string, pipelineDevice string, pipelineSrc string) (*Pipeline, error) {
	pipelineStr := " ! appsink name=appsink"

	var clockRate float32

	switch codecName {
	case webrtc.VP8:
		// https://gstreamer.freedesktop.org/documentation/vpx/vp8enc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1
		if err := CheckPlugins([]string{"ximagesrc", "vpx"}); err != nil {
			return nil, err
		}

		clockRate = videoClockRate

		if pipelineSrc != "" {
			pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		} else {
			pipelineStr = fmt.Sprintf(videoSrc+"vp8enc cpu-used=8 threads=2 deadline=1 error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true"+pipelineStr, pipelineDevice)
		}
	case webrtc.VP9:
		// https://gstreamer.freedesktop.org/documentation/vpx/vp9enc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// vp9enc
		if err := CheckPlugins([]string{"ximagesrc", "vpx"}); err != nil {
			return nil, err
		}

		clockRate = videoClockRate

		// Causes panic! not sure why...
		if pipelineSrc != "" {
			pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		} else {
			pipelineStr = fmt.Sprintf(videoSrc+"vp9enc"+pipelineStr, pipelineDevice)
		}
	case webrtc.H264:
		// https://gstreamer.freedesktop.org/documentation/openh264/openh264enc.html?gi-language=c#openh264enc
		// gstreamer1.0-plugins-bad
		// openh264enc multi-thread=4 complexity=high bitrate=3072000 max-bitrate=4096000
		if err := CheckPlugins([]string{"ximagesrc"}); err != nil {
			return nil, err
		}

		clockRate = videoClockRate

		if pipelineSrc != "" {
			pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		} else {
			pipelineStr = fmt.Sprintf(videoSrc+"openh264enc multi-thread=4 complexity=high bitrate=3072000 max-bitrate=4096000 ! video/x-h264,stream-format=byte-stream"+pipelineStr, pipelineDevice)

			// https://gstreamer.freedesktop.org/documentation/x264/index.html?gi-language=c
			// gstreamer1.0-plugins-ugly
			// video/x-raw,format=I420 ! x264enc bframes=0 key-int-max=60 byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream
			if err := CheckPlugins([]string{"openh264"}); err != nil {
				pipelineStr = fmt.Sprintf(videoSrc+"video/x-raw,format=I420 ! x264enc bframes=0 key-int-max=60 byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream"+pipelineStr, pipelineDevice)

				if err := CheckPlugins([]string{"x264"}); err != nil {
					return nil, err
				}
			}
		}
	case webrtc.Opus:
		// https://gstreamer.freedesktop.org/documentation/opus/opusenc.html
		// gstreamer1.0-plugins-base
		// opusenc
		if err := CheckPlugins([]string{"pulseaudio", "opus"}); err != nil {
			return nil, err
		}

		clockRate = audioClockRate

		if pipelineSrc != "" {
			pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		} else {
			pipelineStr = fmt.Sprintf(audioSrc+"opusenc"+pipelineStr, pipelineDevice)
		}
	case webrtc.G722:
		// https://gstreamer.freedesktop.org/documentation/libav/avenc_g722.html?gi-language=c
		// gstreamer1.0-libav
		// avenc_g722
		if err := CheckPlugins([]string{"pulseaudio", "libav"}); err != nil {
			return nil, err
		}

		clockRate = audioClockRate

		if pipelineSrc != "" {
			pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		} else {
			pipelineStr = fmt.Sprintf(audioSrc+"avenc_g722"+pipelineStr, pipelineDevice)
		}
	case webrtc.PCMU:
		// https://gstreamer.freedesktop.org/documentation/mulaw/mulawenc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// audio/x-raw, rate=8000 ! mulawenc
		if err := CheckPlugins([]string{"pulseaudio", "mulaw"}); err != nil {
			return nil, err
		}

		clockRate = pcmClockRate

		if pipelineSrc != "" {
			pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		} else {
			pipelineStr = fmt.Sprintf(audioSrc+"audio/x-raw, rate=8000 ! mulawenc"+pipelineStr, pipelineDevice)
		}
	case webrtc.PCMA:
		// https://gstreamer.freedesktop.org/documentation/alaw/alawenc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// audio/x-raw, rate=8000 ! alawenc
		if err := CheckPlugins([]string{"pulseaudio", "alaw"}); err != nil {
			return nil, err
		}

		clockRate = pcmClockRate

		if pipelineSrc != "" {
			pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		} else {
			pipelineStr = fmt.Sprintf(audioSrc+"audio/x-raw, rate=8000 ! alawenc"+pipelineStr, pipelineDevice)
		}
	default:
		return nil, fmt.Errorf("unknown codec %s", codecName)
	}

	return CreatePipeline(pipelineStr, codecName, clockRate)
}

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(pipelineStr string, codecName string, clockRate float32) (*Pipeline, error) {
	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	p := &Pipeline{
		Pipeline:  C.gstreamer_send_create_pipeline(pipelineStrUnsafe),
		Sample:    make(chan types.Sample),
		CodecName: codecName,
		ClockRate: clockRate,
		Src:       pipelineStr,
		id:        len(pipelines),
	}

	pipelines[p.id] = p
	return p, nil
}

// Start starts the GStreamer Pipeline
func (p *Pipeline) Start() {
	C.gstreamer_send_start_pipeline(p.Pipeline, C.int(p.id))
}

// Stop stops the GStreamer Pipeline
func (p *Pipeline) Stop() {
	C.gstreamer_send_stop_pipeline(p.Pipeline)
}

// gst-inspect-1.0
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
