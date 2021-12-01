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

	"demodesk/neko/internal/types"
)

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

func CreatePipeline(pipelineStr string) (*Pipeline, error) {
	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	var gstPipeline *C.GstElement
	var gstError *C.GError

	gstPipeline = C.gstreamer_pipeline_create(pipelineStrUnsafe, &gstError)

	if gstError != nil {
		defer C.g_error_free(gstError)
		return nil, fmt.Errorf("(pipeline error) %s", C.GoString(gstError.message))
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

func (p *Pipeline) AttachAppsink(sinkName string) {
	sinkNameUnsafe := C.CString(sinkName)
	defer C.free(unsafe.Pointer(sinkNameUnsafe))

	C.gstreamer_pipeline_attach_appsink(p.Pipeline, sinkNameUnsafe, C.int(p.id))
}

func (p *Pipeline) Play() {
	C.gstreamer_pipeline_play(p.Pipeline)
}

func (p *Pipeline) Stop() {
	C.gstreamer_pipeline_stop(p.Pipeline)
}

func (p *Pipeline) Push(srcName string, buffer []byte) {
	srcNameUnsafe := C.CString(srcName)
	defer C.free(unsafe.Pointer(srcNameUnsafe))

	bytes := C.CBytes(buffer)
	defer C.free(bytes)

	C.gstreamer_pipeline_push(p.Pipeline, srcNameUnsafe, bytes, C.int(len(buffer)))
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
			Data:     C.GoBytes(buffer, bufferLen),
			Duration: time.Duration(duration),
		}
	} else {
		fmt.Printf("discarding buffer, no pipeline with id %d", int(pipelineID))
	}
}
