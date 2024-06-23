package pionlog

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

type logger struct {
	logger    zerolog.Logger
	subsystem string
}

func (l logger) Trace(msg string) {
	l.logger.Trace().Msg(strings.TrimSpace(msg))
}

func (l logger) Tracef(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.logger.Trace().Msg(strings.TrimSpace(msg))
}

func (l logger) Debug(msg string) {
	l.logger.Debug().Msg(strings.TrimSpace(msg))
}

func (l logger) Debugf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.logger.Debug().Msg(strings.TrimSpace(msg))
}

func (l logger) Info(msg string) {
	if strings.Contains(msg, "duplicated packet") {
		return
	}

	l.logger.Info().Msg(strings.TrimSpace(msg))
}

func (l logger) Infof(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if strings.Contains(msg, "duplicated packet") {
		return
	}

	l.logger.Info().Msg(strings.TrimSpace(msg))
}

func (l logger) Warn(msg string) {
	l.logger.Warn().Msg(strings.TrimSpace(msg))
}

func (l logger) Warnf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.logger.Warn().Msg(strings.TrimSpace(msg))
}

func (l logger) Error(msg string) {
	l.logger.Error().Msg(strings.TrimSpace(msg))
}

func (l logger) Errorf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.logger.Error().Msg(strings.TrimSpace(msg))
}
