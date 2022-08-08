package logger

import (
	"github.com/rs/zerolog"
	"os"
)

type Logger struct {
	logger *zerolog.Logger
}

func New() *zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zLog := zerolog.New(os.Stdout).With().Timestamp().Str("service_name", "auth").Logger()
	l := &Logger{logger: &zLog}

	return l.logger
}
