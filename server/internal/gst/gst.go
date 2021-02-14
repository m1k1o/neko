package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include "gst.h"

*/
import "C"
import (
	"fmt"
	"strconv"
	"sync"
	"time"
	"unsafe"

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
	Src       string
	id        int
}

var pipelines = make(map[int]*Pipeline)
var pipelinesLock sync.Mutex
var registry *C.GstRegistry

const (
	videoSrc = "ximagesrc display-name=%s show-pointer=true use-damage=false ! video/x-raw ! videoconvert ! queue ! "
	audioSrc = "pulsesrc device=%s ! audio/x-raw,channels=2 ! audioconvert ! "
)

func init() {
	C.gstreamer_init()
	registry = C.gst_registry_get()
}

// CreateRTMPPipeline creates a GStreamer Pipeline
func CreateRTMPPipeline(pipelineDevice string, pipelineDisplay string, pipelineSrc string, pipelineRTMP string) (*Pipeline, error) {
	video := fmt.Sprintf(videoSrc, pipelineDisplay)
	audio := fmt.Sprintf(audioSrc, pipelineDevice)

	var pipelineStr string
	if pipelineSrc != "" {
		pipelineStr = fmt.Sprintf(pipelineSrc, pipelineRTMP, pipelineDevice, pipelineDisplay)
	} else {
		pipelineStr = fmt.Sprintf("flvmux name=mux ! rtmpsink location='%s live=1' %s audio/x-raw,channels=2 ! audioconvert ! voaacenc ! mux. %s x264enc bframes=0 key-int-max=60 byte-stream=true tune=zerolatency speed-preset=veryfast ! mux.", pipelineRTMP, audio, video)
	}

	return CreatePipeline(pipelineStr, "")
}

// CreateAppPipeline creates a GStreamer Pipeline
func CreateAppPipeline(codecName string, pipelineDevice string, pipelineSrc string, bitrate int) (*Pipeline, error) {
	pipelineStr := " ! appsink name=appsink"

	switch codecName {
	case "VP8":
		// https://gstreamer.freedesktop.org/documentation/vpx/vp8enc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1
		if err := CheckPlugins([]string{"ximagesrc", "vpx"}); err != nil {
			return nil, err
		}

		if pipelineSrc != "" {
			pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		} else {
			pipelineStr = fmt.Sprintf(videoSrc+"vp8enc cpu-used=-5 threads=4 deadline=1 error-resilient=partitions keyframe-max-dist=30 auto-alt-ref=true"+pipelineStr, pipelineDevice)
		}
	case "VP9":
		// https://gstreamer.freedesktop.org/documentation/vpx/vp9enc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// vp9enc
		if err := CheckPlugins([]string{"ximagesrc", "vpx"}); err != nil {
			return nil, err
		}

		// Causes panic! not sure why...
		if pipelineSrc != "" {
			pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		} else {
			pipelineStr = fmt.Sprintf(videoSrc+"vp9enc"+pipelineStr, pipelineDevice)
		}
	case "H264":
		// https://gstreamer.freedesktop.org/documentation/openh264/openh264enc.html?gi-language=c#openh264enc
		// gstreamer1.0-plugins-bad
		// openh264enc multi-thread=4 complexity=high bitrate=3072000 max-bitrate=4096000
		if err := CheckPlugins([]string{"ximagesrc"}); err != nil {
			return nil, err
		}

		if pipelineSrc != "" {
			pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		} else {
			var h264Str string
			 h264Str = "openh264enc multi-thread=4 complexity=high bitrate=3072000 max-bitrate=4096000 ! video/x-h264,stream-format=byte-stream"
			if bitrate > 0 {
				h264Str = "openh264enc multi-thread=4 complexity=high bitrate=" + strconv.Itoa(bitrate) + "000 max-bitrate=" + strconv.Itoa(bitrate) + "999 ! video/x-h264,stream-format=byte-stream"
			}

			// https://gstreamer.freedesktop.org/documentation/x264/index.html?gi-language=c
			// gstreamer1.0-plugins-ugly
			// video/x-raw,format=I420 ! x264enc bframes=0 key-int-max=60 byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream
			if err := CheckPlugins([]string{"openh264"}); err != nil {

				h264Str = "video/x-raw,format=I420 ! x264enc threads=4 byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream"
				if bitrate > 0 {
					h264Str = "video/x-raw,format=I420 ! x264enc threads=4 bitrate=" + strconv.Itoa(bitrate) + " byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream"
				}

				if err := CheckPlugins([]string{"x264"}); err != nil {
					return nil, err
				}
			}
      pipelineStr = fmt.Sprintf(videoSrc+h264Str+pipelineStr, pipelineDevice)
		}
	case "Opus":
		// https://gstreamer.freedesktop.org/documentation/opus/opusenc.html
		// gstreamer1.0-plugins-base
		// opusenc
		if err := CheckPlugins([]string{"pulseaudio", "opus"}); err != nil {
			return nil, err
		}

		if pipelineSrc != "" {
			pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		} else {
			pipelineStr = fmt.Sprintf(audioSrc+"opusenc"+pipelineStr, pipelineDevice)
		}
	case "G722":
		// https://gstreamer.freedesktop.org/documentation/libav/avenc_g722.html?gi-language=c
		// gstreamer1.0-libav
		// avenc_g722
		if err := CheckPlugins([]string{"pulseaudio", "libav"}); err != nil {
			return nil, err
		}

		if pipelineSrc != "" {
			pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		} else {
			pipelineStr = fmt.Sprintf(audioSrc+"avenc_g722"+pipelineStr, pipelineDevice)
		}
	case "PCMU":
		// https://gstreamer.freedesktop.org/documentation/mulaw/mulawenc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// audio/x-raw, rate=8000 ! mulawenc
		if err := CheckPlugins([]string{"pulseaudio", "mulaw"}); err != nil {
			return nil, err
		}

		if pipelineSrc != "" {
			pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		} else {
			pipelineStr = fmt.Sprintf(audioSrc+"audio/x-raw, rate=8000 ! mulawenc"+pipelineStr, pipelineDevice)
		}
	case "PCMA":
		// https://gstreamer.freedesktop.org/documentation/alaw/alawenc.html?gi-language=c
		// gstreamer1.0-plugins-good
		// audio/x-raw, rate=8000 ! alawenc
		if err := CheckPlugins([]string{"pulseaudio", "alaw"}); err != nil {
			return nil, err
		}

		if pipelineSrc != "" {
			pipelineStr = fmt.Sprintf(pipelineSrc+pipelineStr, pipelineDevice)
		} else {
			pipelineStr = fmt.Sprintf(audioSrc+"audio/x-raw, rate=8000 ! alawenc"+pipelineStr, pipelineDevice)
		}
	default:
		return nil, fmt.Errorf("unknown codec %s", codecName)
	}

	return CreatePipeline(pipelineStr, codecName)
}

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(pipelineStr string, codecName string) (*Pipeline, error) {
	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	p := &Pipeline{
		Pipeline:  C.gstreamer_send_create_pipeline(pipelineStrUnsafe),
		Sample:    make(chan types.Sample),
		CodecName: codecName,
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

// Play starts the GStreamer Pipeline
func (p *Pipeline) Play() {
	C.gstreamer_send_play_pipeline(p.Pipeline)
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
		pipeline.Sample <- types.Sample{Data: C.GoBytes(buffer, bufferLen), Timestamp: time.Now(), Duration: time.Duration(duration)}
	} else {
		fmt.Printf("discarding buffer, no pipeline with id %d", int(pipelineID))
	}
	C.free(buffer)
}
