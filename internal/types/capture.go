package types

import (
	"math"
	"strings"
	"fmt"

	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/PaesslerAG/gval"

	"demodesk/neko/internal/types/codec"
)

type Sample media.Sample

type BroadcastManager interface {
	Start(url string) error
	Stop()
	Started() bool
	Url() string
}

type ScreencastManager interface {
	Enabled() bool
	Started() bool
	Image() ([]byte, error)
}

type StreamManager interface {
	Codec() codec.RTPCodec

	AddListener(listener *func(sample Sample))
	RemoveListener(listener *func(sample Sample))
	ListenersCount() int

	Start() error
	Stop()
	Started() bool
}

type CaptureManager interface {
	Start()
	Shutdown() error

	Broadcast() BroadcastManager
	Screencast() ScreencastManager
	Audio() StreamManager
	Video(videoID string) (StreamManager, bool)
	VideoIDs() []string
}

type VideoConfig struct {
	Codec       string            `mapstructure:"codec"`
	Width       string            `mapstructure:"width"`        // expression
	Height      string            `mapstructure:"height"`       // expression
	Fps         string            `mapstructure:"fps"`          // expression
	GstEncoder  string            `mapstructure:"gst_encoder"`
	GstParams   map[string]string `mapstructure:"gst_params"`   // map of expressions
	GstPipeline string            `mapstructure:"gst_pipeline"`
}

func (config *VideoConfig) GetCodec() (codec.RTPCodec, error) {
	switch strings.ToLower(config.Codec) {
	case "vp8":
		return codec.VP8(), nil
	case "vp9":
		return codec.VP9(), nil
	case "h264":
		return codec.H264(), nil
	default:
		return codec.RTPCodec{}, fmt.Errorf("unknown codec")
	}
}

func (config *VideoConfig) GetPipeline(screen ScreenSize) (string, error) {
	if config.GstPipeline != "" {
		return config.GstPipeline, nil
	}

	values := map[string]interface{}{
		"width":  screen.Width,
		"height": screen.Height,
		"fps":    screen.Rate,
	}

	language := []gval.Language{
		gval.Function("round", func(args ...interface{}) (interface{}, error) {
			return (int)(math.Round(args[0].(float64))), nil
		}),
	}

	// get fps pipeline
	fpsPipeline := "! video/x-raw ! videoconvert ! queue"
	if config.Fps != "" {
		var err error
		val, err := gval.Evaluate(config.Fps, values, language...)
		if err != nil {
			return "", err
		}
	
		if val != nil {
			// TODO: To fraction.
			fpsPipeline = fmt.Sprintf("! video/x-raw,framerate=%v ! videoconvert ! queue", val)
		}
	}

	// get scale pipeline
	scalePipeline := ""
	if config.Width != "" && config.Height != "" {
		w, err := gval.Evaluate(config.Width, values, language...)
		if err != nil {
			return "", err
		}

		h, err := gval.Evaluate(config.Height, values, language...)
		if err != nil {
			return "", err
		}

		if w != nil && h != nil {
			scalePipeline = fmt.Sprintf("! videoscale ! video/x-raw,width=%v,height=%v ! queue", w, h)
		}
	}

	// get encoder pipeline
	encPipeline := fmt.Sprintf("! %s", config.GstEncoder)
	for key, expr := range config.GstParams {
		if expr == "" {
			continue
		}

		val, err := gval.Evaluate(expr, values, language...)
		if err != nil {
			return "", err
		}

		if val != nil {
			encPipeline += fmt.Sprintf(" %s=%v", key, val)
		} else {
			encPipeline += fmt.Sprintf(" %s=%s", key, expr)
		}
	}

	return fmt.Sprintf("%s %s %s", fpsPipeline, scalePipeline, encPipeline), nil
}
