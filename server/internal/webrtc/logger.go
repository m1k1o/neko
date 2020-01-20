package webrtc

import (
	"github.com/pion/logging"
	"github.com/rs/zerolog"
)

type logger struct {
	logger zerolog.Logger
}

func (l logger) Trace(msg string)                          { l.logger.Trace().Msg(msg) }
func (l logger) Tracef(format string, args ...interface{}) { l.logger.Trace().Msgf(format, args...) }
func (l logger) Debug(msg string)                          { l.logger.Debug().Msg(msg) }
func (l logger) Debugf(format string, args ...interface{}) { l.logger.Debug().Msgf(format, args...) }
func (l logger) Info(msg string)                           { l.logger.Info().Msg(msg) }
func (l logger) Infof(format string, args ...interface{})  { l.logger.Info().Msgf(format, args...) }
func (l logger) Warn(msg string)                           { l.logger.Warn().Msg(msg) }
func (l logger) Warnf(format string, args ...interface{})  { l.logger.Warn().Msgf(format, args...) }
func (l logger) Error(msg string)                          { l.logger.Error().Msg(msg) }
func (l logger) Errorf(format string, args ...interface{}) { l.logger.Error().Msgf(format, args...) }

type loggerFactory struct {
	logger zerolog.Logger
}

func (l loggerFactory) NewLogger(subsystem string) logging.LeveledLogger {
	return logger{
		logger: l.logger.With().Str("subsystem", subsystem).Logger(),
	}
}
