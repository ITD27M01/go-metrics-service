package logging

import (
	"net/http"
	"strings"

	"github.com/go-chi/httplog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogLevel sets level of logging
func LogLevel(level string) {
	switch level {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "WARNING":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		log.Fatal().Msgf("Unsupported log Level: %s", level)
	}
}

// HTTPRequestLogger is a special middleware for logging HTTP requests
func HTTPRequestLogger(level string) func(next http.Handler) http.Handler {
	logger := httplog.NewLogger("metrics", httplog.Options{
		JSON:     false,
		LogLevel: strings.ToLower(level),
	})

	return httplog.RequestLogger(logger)
}
