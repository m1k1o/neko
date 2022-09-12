package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include "gst.h"
*/
import "C"
import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"m1k1o/neko/internal/types"
)

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	Pipeline *C.GstElement
	Sample   chan types.Sample
	Src      string
	id       int
}

var pipelines = make(map[int]*Pipeline)
var pipelinesLock sync.Mutex
var registry *C.GstRegistry

func init() {
	C.gstreamer_init()
	registry = C.gst_registry_get()
}

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(pipelineStr string) (*Pipeline, error) {
	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	var err *C.GError
	gstPipeline := C.gstreamer_send_create_pipeline(pipelineStrUnsafe, &err)
	if err != nil {
		defer C.g_error_free(err)
		return nil, fmt.Errorf("%s", C.GoString(err.message))
	}

	p := &Pipeline{
		Pipeline: gstPipeline,
		Sample:   make(chan types.Sample),
		Src:      pipelineStr,
		id:       len(pipelines),
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
