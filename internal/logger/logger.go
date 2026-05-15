package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func NewLogger(app string, level string) *zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	l := zerolog.New(os.Stdout).With().Timestamp().Logger().With().Str("app", app).Logger().Level(zerolog.InfoLevel)

	parsed, err := zerolog.ParseLevel(level)

	if err != nil {
		return &l
	}

	l = l.Level(parsed)

	return &l
}
