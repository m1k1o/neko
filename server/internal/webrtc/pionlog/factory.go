package pionlog

import (
	"github.com/pion/logging"
	"github.com/rs/zerolog"
)

func New(logger zerolog.Logger) Factory {
	return Factory{
		Logger: logger.With().Str("submodule", "pion").Logger(),
	}
}

type Factory struct {
	Logger zerolog.Logger
}

func (l Factory) NewLogger(subsystem string) logging.LeveledLogger {
	if subsystem == "sctp" {
		return nulllog{}
	}

	return logger{
		subsystem: subsystem,
		logger:    l.Logger.With().Str("subsystem", subsystem).Logger(),
	}
}
