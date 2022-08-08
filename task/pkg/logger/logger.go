package logger

import (
	"os"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger *zerolog.Logger
}

func New() *zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zLog := zerolog.New(os.Stdout).With().Timestamp().Str("service_name", "task").Logger()
	l := &Logger{logger: &zLog}

	return l.logger
}
