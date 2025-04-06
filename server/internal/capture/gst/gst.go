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

	"m1k1o/neko/internal/types"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
)

type Pipeline struct {
	id     int
	logger zerolog.Logger
	Src    string
	Ctx    *C.GstPipelineCtx
	Sample chan types.Sample
}

var pSerial int32
var pipelines = make(map[int]*Pipeline)
var pipelinesLock sync.Mutex
var registry *C.GstRegistry
var gMainLoop *C.GMainLoop

func init() {
	gst.InitCheck()
	registry = C.gst_registry_get()
}

func RunMainLoop() {
	if gMainLoop != nil {
		return
	}
	gMainLoop = C.g_main_loop_new(nil, C.int(0))
	C.g_main_loop_run(gMainLoop)
}

func QuitMainLoop() {
	if gMainLoop == nil {
		return
	}
	C.g_main_loop_quit(gMainLoop)
	gMainLoop = nil
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
		fmt.Printf("(pipeline error) %s", C.GoString(gstError.message))
		return nil, fmt.Errorf("(pipeline error) %s", C.GoString(gstError.message))
	}

	p := &Pipeline{
		id: int(id),
		logger: log.With().
			Str("module", "capture").
			Str("submodule", "gstreamer").
			Int("pipeline_id", int(id)).Logger(),
		Src: pipelineStr,
		Ctx: ctx,
	}

	pipelines[p.id] = p
	return p, nil
}

func (p *Pipeline) AttachAppsink(sinkName string, sampleChannel chan types.Sample) {
	sinkNameUnsafe := C.CString(sinkName)
	defer C.free(unsafe.Pointer(sinkNameUnsafe))

	p.Sample = sampleChannel

	C.gstreamer_pipeline_attach_appsink(p.Ctx, sinkNameUnsafe)
}

func (p *Pipeline) AttachAppsrc(srcName string) {
	srcNameUnsafe := C.CString(srcName)
	defer C.free(unsafe.Pointer(srcNameUnsafe))

	C.gstreamer_pipeline_attach_appsrc(p.Ctx, srcNameUnsafe)
}

func (p *Pipeline) Play() {
	p.logger.Info().Msg("starting pipeline")
	C.gstreamer_pipeline_play(p.Ctx)
}

func (p *Pipeline) Stop() {
	p.logger.Info().Msg("stopping and destroying pipeline")
	C.gstreamer_pipeline_destory(p.Ctx)

	pipelinesLock.Lock()
	delete(pipelines, p.id)
	pipelinesLock.Unlock()
}

func (p *Pipeline) Pause() {
	C.gstreamer_pipeline_pause(p.Ctx)
}

func (p *Pipeline) Destroy() {
	p.logger.Info().Msg("destroying pipeline resources")
	C.gstreamer_pipeline_destory(p.Ctx)

	pipelinesLock.Lock()
	delete(pipelines, p.id)
	pipelinesLock.Unlock()

	if p.Ctx != nil {
		C.free(unsafe.Pointer(p.Ctx))
		p.Ctx = nil
	}
}

func (p *Pipeline) Push(buffer []byte) {
	bytes := C.CBytes(buffer)
	defer C.free(bytes)

	C.gstreamer_pipeline_push(p.Ctx, bytes, C.int(len(buffer)))
}

func (p *Pipeline) SetPropInt(binName string, prop string, value int) bool {
	cBinName := C.CString(binName)
	defer C.free(unsafe.Pointer(cBinName))

	cProp := C.CString(prop)
	defer C.free(unsafe.Pointer(cProp))

	cValue := C.int(value)

	p.logger.Debug().Msgf("setting prop %s of %s to %d", prop, binName, value)

	ok := C.gstreamer_pipeline_set_prop_int(p.Ctx, cBinName, cProp, cValue)
	return ok == C.TRUE
}

func (p *Pipeline) SetCapsFramerate(binName string, numerator, denominator int) bool {
	cBinName := C.CString(binName)
	cNumerator := C.int(numerator)
	cDenominator := C.int(denominator)

	defer C.free(unsafe.Pointer(cBinName))

	p.logger.Debug().Msgf("setting caps framerate of %s to %d/%d", binName, numerator, denominator)

	ok := C.gstreamer_pipeline_set_caps_framerate(p.Ctx, cBinName, cNumerator, cDenominator)
	return ok == C.TRUE
}

func (p *Pipeline) SetCapsResolution(binName string, width, height int) bool {
	cBinName := C.CString(binName)
	cWidth := C.int(width)
	cHeight := C.int(height)

	defer C.free(unsafe.Pointer(cBinName))

	p.logger.Debug().Msgf("setting caps resolution of %s to %dx%d", binName, width, height)

	ok := C.gstreamer_pipeline_set_caps_resolution(p.Ctx, cBinName, cWidth, cHeight)
	return ok == C.TRUE
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
			Data:      C.GoBytes(buffer, bufferLen),
			Timestamp: time.Now(),
			Duration:  time.Duration(duration),
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

// GetAppSink retrieves the appsink element stored in the C context.
// Returns nil if the appsink was not found/stored or casting fails.
func (p *Pipeline) GetAppSink() *app.Sink {
	element := p.Ctx.appsink
	if element == nil {
		p.logger.Warn().Msg("GetAppSink: C context appsink field is nil")
		return nil
	}

	gstElement := gst.WrapElement(unsafe.Pointer(element))
	if gstElement == nil {
		p.logger.Error().Msg("GetAppSink: failed to wrap GstElement from C context")
		return nil
	}

	appSink, ok := gstElement.Cast().(*app.Sink)
	if !ok || appSink == nil {
		p.logger.Warn().Msg("GetAppSink: element from C context is not an app.Sink")
		if bin, okBin := gstElement.Instance().(*gst.Bin); okBin {
			if appSinkBin, okAppSink := bin.Cast().(*app.Sink); okAppSink {
				p.logger.Debug().Msg("GetAppSink: Cast via Bin successful")
				return appSinkBin
			}
		}
		p.logger.Error().Msg("GetAppSink: Failed to cast element to app.Sink even via Bin")
		return nil
	}

	return appSink
}
