package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include "gst.h"
*/
import "C"
import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/demodesk/neko/pkg/types"
)

type Pipeline struct {
	id     int
	Src    string
	Ctx    *C.GstPipelineCtx
	Sample chan types.Sample
}

var pSerial int32
var pipelines = make(map[int]*Pipeline)
var pipelinesLock sync.Mutex
var registry *C.GstRegistry

func init() {
	C.gstreamer_init()
	go C.gstreamer_loop()

	registry = C.gst_registry_get()
}

func CreatePipeline(pipelineStr string) (*Pipeline, error) {
	id := atomic.AddInt32(&pSerial, 1)

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	var gstError *C.GError
	ctx := C.gstreamer_pipeline_create(pipelineStrUnsafe, C.int(id), &gstError)

	if gstError != nil {
		defer C.g_error_free(gstError)
		return nil, fmt.Errorf("(pipeline error) %s", C.GoString(gstError.message))
	}

	p := &Pipeline{
		id:     int(id),
		Src:    pipelineStr,
		Ctx:    ctx,
		Sample: make(chan types.Sample),
	}

	pipelines[p.id] = p
	return p, nil
}

func (p *Pipeline) AttachAppsink(sinkName string) {
	sinkNameUnsafe := C.CString(sinkName)
	defer C.free(unsafe.Pointer(sinkNameUnsafe))

	C.gstreamer_pipeline_attach_appsink(p.Ctx, sinkNameUnsafe)
}

func (p *Pipeline) AttachAppsrc(srcName string) {
	srcNameUnsafe := C.CString(srcName)
	defer C.free(unsafe.Pointer(srcNameUnsafe))

	C.gstreamer_pipeline_attach_appsrc(p.Ctx, srcNameUnsafe)
}

func (p *Pipeline) Play() {
	C.gstreamer_pipeline_play(p.Ctx)
}

func (p *Pipeline) Pause() {
	C.gstreamer_pipeline_pause(p.Ctx)
}

func (p *Pipeline) Destroy() {
	C.gstreamer_pipeline_destory(p.Ctx)

	pipelinesLock.Lock()
	delete(pipelines, p.id)
	pipelinesLock.Unlock()

	close(p.Sample)
	C.free(unsafe.Pointer(p.Ctx))
	p = nil
}

func (p *Pipeline) Push(buffer []byte) {
	bytes := C.CBytes(buffer)
	defer C.free(bytes)

	C.gstreamer_pipeline_push(p.Ctx, bytes, C.int(len(buffer)))
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
		log.Warn().
			Str("module", "capture").
			Str("submodule", "gstreamer").
			Int("pipeline_id", int(pipelineID)).
			Msgf("discarding sample, pipeline not found")
	}
}

//export goPipelineLog
func goPipelineLog(levelUnsafe *C.char, msgUnsafe *C.char, pipelineID C.int) {
	levelStr := C.GoString(levelUnsafe)
	msg := C.GoString(msgUnsafe)

	level, _ := zerolog.ParseLevel(levelStr)
	log.WithLevel(level).
		Str("module", "capture").
		Str("submodule", "gstreamer").
		Int("pipeline_id", int(pipelineID)).
		Msg(msg)
}
