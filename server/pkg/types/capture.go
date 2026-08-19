package types

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/m1k1o/neko/server/pkg/types/codec"

	"github.com/PaesslerAG/gval"
)

var (
	ErrCapturePipelineAlreadyExists = errors.New("capture pipeline already exists")
)

type Sample struct {
	// timing information
	Timestamp time.Time
	// PTS, DTS, and Duration use -1 for GST_CLOCK_TIME_NONE.
	PTS      time.Duration
	DTS      time.Duration
	Duration time.Duration
	// metadata
	DeltaUnit bool // this unit cannot be decoded independently.
	// buffer length
	Length int
	// buffer with encoded media
	Data []byte
}

type SampleConsumer interface {
	WriteSample(Sample)
}

type EncodedStream interface {
	ID() string
	Codec() codec.RTPCodec
	Bitrate() uint64

	Subscribe(SampleConsumer) (StreamSubscription, error)
}

type StreamSubscription interface {
	Stream() EncodedStream
	Switch(EncodedStream) error
	Close() error
}

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

type StreamSelectorType int

const (
	// select exact stream
	StreamSelectorTypeExact StreamSelectorType = iota
	// select nearest stream (in either direction) if exact stream is not available
	StreamSelectorTypeNearest
	// if exact stream is found select the next lower stream, otherwise select the nearest lower stream
	StreamSelectorTypeLower
	// if exact stream is found select the next higher stream, otherwise select the nearest higher stream
	StreamSelectorTypeHigher
)

func (s StreamSelectorType) String() string {
	switch s {
	case StreamSelectorTypeExact:
		return "exact"
	case StreamSelectorTypeNearest:
		return "nearest"
	case StreamSelectorTypeLower:
		return "lower"
	case StreamSelectorTypeHigher:
		return "higher"
	default:
		return fmt.Sprintf("%d", int(s))
	}
}

func (s *StreamSelectorType) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "exact", "":
		*s = StreamSelectorTypeExact
	case "nearest":
		*s = StreamSelectorTypeNearest
	case "lower":
		*s = StreamSelectorTypeLower
	case "higher":
		*s = StreamSelectorTypeHigher
	default:
		return fmt.Errorf("invalid stream selector type: %s", string(text))
	}
	return nil
}

func (s StreamSelectorType) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

type StreamSelector struct {
	// type of stream selector
	Type StreamSelectorType `json:"type"`
	// select stream by its ID
	ID string `json:"id"`
	// select stream by its bitrate
	Bitrate uint64 `json:"bitrate"`
}

type EncodedStreamSelector interface {
	IDs() []string
	Codec() codec.RTPCodec

	GetStream(selector StreamSelector) (EncodedStream, bool)
}

type EncodedMediaSource interface {
	Audio() EncodedStream
	Video() EncodedStreamSelector
}

type StreamSrcManager interface {
	Codec() codec.RTPCodec

	Start(codec codec.RTPCodec) error
	Stop()
	Push(bytes []byte)

	Started() bool
}

type CaptureManager interface {
	EncodedMediaSource

	Start()
	Shutdown() error

	Broadcast() BroadcastManager
	Screencast() ScreencastManager

	Webcam() StreamSrcManager
	Microphone() StreamSrcManager
}

type VideoConfig struct {
	Width       string            `mapstructure:"width"`        // expression
	Height      string            `mapstructure:"height"`       // expression
	Fps         string            `mapstructure:"fps"`          // expression
	Bitrate     int               `mapstructure:"bitrate"`      // pipeline bitrate (not used currently)
	GstPrefix   string            `mapstructure:"gst_prefix"`   // pipeline prefix, starts with !
	GstEncoder  string            `mapstructure:"gst_encoder"`  // gst encoder name
	GstParams   map[string]string `mapstructure:"gst_params"`   // map of expressions
	GstSuffix   string            `mapstructure:"gst_suffix"`   // pipeline suffix, starts with !
	GstPipeline string            `mapstructure:"gst_pipeline"` // whole pipeline as a string
	ShowPointer bool              `mapstructure:"show_pointer"` // show pointer in the video
}

func (config *VideoConfig) GetPipeline(screen ScreenSize) (string, error) {
	values := map[string]any{
		"width":  screen.Width,
		"height": screen.Height,
		"fps":    screen.Rate,
	}

	language := []gval.Language{
		gval.Function("round", func(args ...any) (any, error) {
			return (int)(math.Round(args[0].(float64))), nil
		}),
	}

	// get fps pipeline
	fpsPipeline := "! video/x-raw ! videoconvert ! queue"
	if config.Fps != "" {
		eval, err := gval.Full(language...).NewEvaluable(config.Fps)
		if err != nil {
			return "", err
		}

		val, err := eval.EvalFloat64(context.Background(), values)
		if err != nil {
			return "", err
		}

		fpsPipeline = fmt.Sprintf("! capsfilter caps=video/x-raw,framerate=%d/100 name=framerate ! videoconvert ! queue", int(val*100))
	}

	// get scale pipeline
	scalePipeline := ""
	if config.Width != "" && config.Height != "" {
		eval, err := gval.Full(language...).NewEvaluable(config.Width)
		if err != nil {
			return "", err
		}

		w, err := eval.EvalInt(context.Background(), values)
		if err != nil {
			return "", err
		}

		eval, err = gval.Full(language...).NewEvaluable(config.Height)
		if err != nil {
			return "", err
		}

		h, err := eval.EvalInt(context.Background(), values)
		if err != nil {
			return "", err
		}

		// element videoscale parameter method to 0 meaning nearest neighbor
		scalePipeline = fmt.Sprintf("! videoscale method=0 ! capsfilter caps=video/x-raw,width=%d,height=%d name=resolution ! queue", w, h)
	}

	// get encoder pipeline
	encPipeline := fmt.Sprintf("! %s name=encoder", config.GstEncoder)
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

	// join strings with space
	return strings.Join([]string{
		fpsPipeline,
		scalePipeline,
		config.GstPrefix,
		encPipeline,
		config.GstSuffix,
	}[:], " "), nil
}
