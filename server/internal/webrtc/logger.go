package webrtc

import (
	"fmt"
	"strings"

	"github.com/pion/logging"
	"github.com/rs/zerolog"
)

type nulllog struct{}

func (l nulllog) Trace(msg string)                          {}
func (l nulllog) Tracef(format string, args ...interface{}) {}
func (l nulllog) Debug(msg string)                          {}
func (l nulllog) Debugf(format string, args ...interface{}) {}
func (l nulllog) Info(msg string)                           {}
func (l nulllog) Infof(format string, args ...interface{})  {}
func (l nulllog) Warn(msg string)                           {}
func (l nulllog) Warnf(format string, args ...interface{})  {}
func (l nulllog) Error(msg string)                          {}
func (l nulllog) Errorf(format string, args ...interface{}) {}

type logger struct {
	logger    zerolog.Logger
	subsystem string
}

func (l logger) Trace(msg string)                          { l.logger.Trace().Msg(msg) }
func (l logger) Tracef(format string, args ...interface{}) { l.logger.Trace().Msgf(format, args...) }
func (l logger) Debug(msg string)                          { l.logger.Debug().Msg(msg) }
func (l logger) Debugf(format string, args ...interface{}) { l.logger.Debug().Msgf(format, args...) }
func (l logger) Info(msg string) {
	if strings.Contains(msg, "packetio.Buffer is full") {
		//l.logger.Panic().Msg(msg)
		return
	}
	l.logger.Info().Msg(msg)
}
func (l logger) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if strings.Contains(msg, "packetio.Buffer is full") {
		// l.logger.Panic().Msg(msg)
		return
	}
	l.logger.Info().Msg(msg)
}
func (l logger) Warn(msg string)                           { l.logger.Warn().Msg(msg) }
func (l logger) Warnf(format string, args ...interface{})  { l.logger.Warn().Msgf(format, args...) }
func (l logger) Error(msg string)                          { l.logger.Error().Msg(msg) }
func (l logger) Errorf(format string, args ...interface{}) { l.logger.Error().Msgf(format, args...) }

type loggerFactory struct {
	logger zerolog.Logger
}

func (l loggerFactory) NewLogger(subsystem string) logging.LeveledLogger {
	if subsystem == "sctp" {
		return nulllog{}
	}

	return logger{
		subsystem: subsystem,
		logger:    l.logger.With().Str("subsystem", subsystem).Logger(),
	}
}
