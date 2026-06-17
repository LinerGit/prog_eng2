package logger

import (
	"os"
	"strings"
	"time"

	"ticket/internal/config"

	"github.com/rs/zerolog"
)

func New(cfg *config.Config) zerolog.Logger {
	level, err := zerolog.ParseLevel(strings.ToLower(cfg.LogLevel))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	if cfg.AppEnv == "production" {
		return zerolog.New(os.Stdout).With().
			Timestamp().
			Str("service", cfg.ServiceName).
			Str("env", cfg.AppEnv).
			Logger()
	}

	return zerolog.New(writer).With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Str("env", cfg.AppEnv).
		Logger()
}
