package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include "gst.h"

*/
import "C"
import (
	"fmt"
	"io"
	"sync"
	"unsafe"

	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
)

func init() {
	go C.gstreamer_send_start_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	Pipeline  *C.GstElement
	tracks    []*webrtc.Track
	id        int
	codecName string
	clockRate float32
}

var pipelines = make(map[int]*Pipeline)
var pipelinesLock sync.Mutex

const (
	videoClockRate = 90000
	audioClockRate = 48000
	pcmClockRate   = 8000
)

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(codecName string, tracks []*webrtc.Track, pipelineSrc string) *Pipeline {
	pipelineStr := "appsink name=appsink"
	var clockRate float32

	switch codecName {
	case webrtc.VP8:
		pipelineStr = pipelineSrc + " ! vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1 ! " + pipelineStr
		clockRate = videoClockRate

	case webrtc.VP9:
		pipelineStr = pipelineSrc + " ! vp9enc ! " + pipelineStr
		clockRate = videoClockRate

	case webrtc.H264:
		pipelineStr = pipelineSrc + " ! video/x-raw,format=I420 ! x264enc bframes=0 speed-preset=veryfast key-int-max=60 ! video/x-h264,stream-format=byte-stream ! " + pipelineStr
		clockRate = videoClockRate

	case webrtc.Opus:
		pipelineStr = pipelineSrc + " ! opusenc ! " + pipelineStr
		clockRate = audioClockRate

	case webrtc.G722:
		pipelineStr = pipelineSrc + " ! avenc_g722 ! " + pipelineStr
		clockRate = audioClockRate

	case webrtc.PCMU:
		pipelineStr = pipelineSrc + " ! audio/x-raw, rate=8000 ! mulawenc ! " + pipelineStr
		clockRate = pcmClockRate

	case webrtc.PCMA:
		pipelineStr = pipelineSrc + " ! audio/x-raw, rate=8000 ! alawenc ! " + pipelineStr
		clockRate = pcmClockRate

	default:
		panic("Unhandled codec " + codecName)
	}

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	pipeline := &Pipeline{
		Pipeline:  C.gstreamer_send_create_pipeline(pipelineStrUnsafe),
		tracks:    tracks,
		id:        len(pipelines),
		codecName: codecName,
		clockRate: clockRate,
	}

	pipelines[pipeline.id] = pipeline
	return pipeline
}

// Start starts the GStreamer Pipeline
func (p *Pipeline) Start() {
	C.gstreamer_send_start_pipeline(p.Pipeline, C.int(p.id))
}

// Stop stops the GStreamer Pipeline
func (p *Pipeline) Stop() {
	C.gstreamer_send_stop_pipeline(p.Pipeline)
}

//export goHandlePipelineBuffer
func goHandlePipelineBuffer(buffer unsafe.Pointer, bufferLen C.int, duration C.int, pipelineID C.int) {
	pipelinesLock.Lock()
	pipeline, ok := pipelines[int(pipelineID)]
	pipelinesLock.Unlock()

	if ok {
		samples := uint32(pipeline.clockRate * (float32(duration) / 1000000000))
		for _, t := range pipeline.tracks {
			if err := t.WriteSample(media.Sample{Data: C.GoBytes(buffer, bufferLen), Samples: samples}); err != nil && err != io.ErrClosedPipe {
				panic(err)
			}
		}
	} else {
		fmt.Printf("discarding buffer, no pipeline with id %d", int(pipelineID))
	}
	C.free(buffer)
}
