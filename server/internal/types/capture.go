package types

import (
	"errors"

	"m1k1o/neko/internal/types/codec"
)

var (
	ErrCapturePipelineAlreadyExists = errors.New("capture pipeline already exists")
)

type BroadcastManager interface {
	Start(url string) error
	Stop()
	Started() bool
	Url() string
}

type StreamSinkManager interface {
	Codec() codec.RTPCodec

	AddListener() error
	RemoveListener() error

	ListenersCount() int
	Started() bool
	GetSampleChannel() chan Sample
}

type CaptureManager interface {
	Start()
	Shutdown() error

	Broadcast() BroadcastManager
	Audio() StreamSinkManager
	Video() StreamSinkManager
}
