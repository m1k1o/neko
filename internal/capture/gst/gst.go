package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include "gst.h"
*/
import "C"
import (
	"fmt"
	"time"
	"sync"
	"unsafe"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/codec"
)

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	Pipeline  *C.GstElement
	Sample    chan types.Sample
	Src       string
	id        int
}

var pipelines = make(map[int]*Pipeline)
var pipelinesLock sync.Mutex
var registry *C.GstRegistry

const (
	videoSrc = "ximagesrc display-name=%s show-pointer=false use-damage=false ! video/x-raw ! videoconvert ! queue ! "
	audioSrc = "pulsesrc device=%s ! audio/x-raw,channels=2 ! audioconvert ! "
	appSink  = " ! appsink name=appsink"
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

	return CreatePipeline(pipelineStr)
}

// CreateJPEGPipeline creates a GStreamer Pipeline
func CreateJPEGPipeline(pipelineDisplay string, pipelineSrc string, rate string, quality string) (*Pipeline, error) {
	var pipelineStr string
	if pipelineSrc != "" {
		pipelineStr = fmt.Sprintf(pipelineSrc, pipelineDisplay)
	} else {
		pipelineStr = fmt.Sprintf("ximagesrc display-name=%s show-pointer=true use-damage=false ! videoconvert ! videoscale ! videorate ! video/x-raw,framerate=%s ! jpegenc quality=%s" + appSink, pipelineDisplay, rate, quality)
	}

	return CreatePipeline(pipelineStr)
}

// CreateAppPipeline creates a GStreamer Pipeline
func CreateAppPipeline(codecRTP codec.RTPCodec, pipelineDevice string, pipelineSrc string) (*Pipeline, error) {
	var pipelineStr string

	switch codecRTP.Name {
	case "vp8":
		// https://gstreamer.freedesktop.org/documentation/vpx/vp8enc.html
		// gstreamer1.0-plugins-good
		if err := CheckPlugins([]string{"ximagesrc", "vpx"}); err != nil {
			return nil, err
		}

		pipelineStr = fmt.Sprintf(videoSrc + "vp8enc cpu-used=16 threads=4 deadline=1 error-resilient=partitions keyframe-max-dist=15 static-threshold=20" + appSink, pipelineDevice)
	case "vp9":
		// https://gstreamer.freedesktop.org/documentation/vpx/vp9enc.html
		// gstreamer1.0-plugins-good
		if err := CheckPlugins([]string{"ximagesrc", "vpx"}); err != nil {
			return nil, err
		}

		pipelineStr = fmt.Sprintf(videoSrc + "vp9enc" + appSink, pipelineDevice)
	case "h264":
		var err error
		if err = CheckPlugins([]string{"ximagesrc"}); err != nil {
			return nil, err
		}

		// https://gstreamer.freedesktop.org/documentation/x264/index.html
		// gstreamer1.0-plugins-ugly
		if err = CheckPlugins([]string{"x264"}); err == nil {
			pipelineStr = fmt.Sprintf(videoSrc + "video/x-raw,format=I420 ! x264enc threads=4 bitrate=4096 key-int-max=15 byte-stream=true tune=zerolatency speed-preset=veryfast ! video/x-h264,stream-format=byte-stream" + appSink, pipelineDevice)
			break
		}

		// https://gstreamer.freedesktop.org/documentation/openh264/openh264enc.html
		// gstreamer1.0-plugins-bad
		if err = CheckPlugins([]string{"openh264"}); err == nil {
			pipelineStr = fmt.Sprintf(videoSrc + "openh264enc multi-thread=4 complexity=high bitrate=3072000 max-bitrate=4096000 ! video/x-h264,stream-format=byte-stream" + appSink, pipelineDevice)
			break
		}

		return nil, err
	case "opus":
		// https://gstreamer.freedesktop.org/documentation/opus/opusenc.html
		// gstreamer1.0-plugins-base
		if err := CheckPlugins([]string{"pulseaudio", "opus"}); err != nil {
			return nil, err
		}

		pipelineStr = fmt.Sprintf(audioSrc + "opusenc bitrate=128000" + appSink, pipelineDevice)
	case "g722":
		// https://gstreamer.freedesktop.org/documentation/libav/avenc_g722.html
		// gstreamer1.0-libav
		if err := CheckPlugins([]string{"pulseaudio", "libav"}); err != nil {
			return nil, err
		}

		pipelineStr = fmt.Sprintf(audioSrc + "avenc_g722" + appSink, pipelineDevice)
	case "pcmu":
		// https://gstreamer.freedesktop.org/documentation/mulaw/mulawenc.html
		// gstreamer1.0-plugins-good
		if err := CheckPlugins([]string{"pulseaudio", "mulaw"}); err != nil {
			return nil, err
		}

		pipelineStr = fmt.Sprintf(audioSrc + "audio/x-raw, rate=8000 ! mulawenc" + appSink, pipelineDevice)
	case "pcma":
		// https://gstreamer.freedesktop.org/documentation/alaw/alawenc.html
		// gstreamer1.0-plugins-good
		if err := CheckPlugins([]string{"pulseaudio", "alaw"}); err != nil {
			return nil, err
		}

		pipelineStr = fmt.Sprintf(audioSrc + "audio/x-raw, rate=8000 ! alawenc" + appSink, pipelineDevice)
	default:
		return nil, fmt.Errorf("unknown codec %s", codecRTP.Name)
	}

	if pipelineSrc != "" {
		pipelineStr = fmt.Sprintf(pipelineSrc + appSink, pipelineDevice)
	}

	return CreatePipeline(pipelineStr)
}

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(pipelineStr string) (*Pipeline, error) {
	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	var gstPipeline *C.GstElement
	var gstError *C.GError

	gstPipeline = C.gst_parse_launch(pipelineStrUnsafe, &gstError)

	if gstError != nil {
		defer C.g_error_free(gstError)
		return nil, fmt.Errorf("(pipeline error) %s", C.GoString(gstError.message)) 
	}

	p := &Pipeline{
		Pipeline:  gstPipeline,
		Sample:    make(chan types.Sample),
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
	defer C.free(buffer)

	pipelinesLock.Lock()
	pipeline, ok := pipelines[int(pipelineID)]
	pipelinesLock.Unlock()

	if ok {
		pipeline.Sample <- types.Sample{
			Data: C.GoBytes(buffer, bufferLen),
			Duration: time.Duration(duration),
		}
	} else {
		fmt.Printf("discarding buffer, no pipeline with id %d", int(pipelineID))
	}
}
