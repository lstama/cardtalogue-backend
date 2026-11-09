// Package logging sets up structured logging with the standard library
// log/slog. Log format (json/text) and level are driven by LOG_FORMAT and
// LOG_LEVEL.
//
// OTel log export (OTLP) is intentionally out of scope for Phase 1: the
// otlplog exporters are not available in this environment's module proxy. When
// they become available, a slog->OTel bridge can be wired in here.
package logging

import (
	"log/slog"
	"os"
	"strings"

	"github.com/lstama/cardtalogue-backend/internal/config"
)

// New builds a *slog.Logger from configuration and installs it as the global
// logger. It returns the logger so callers can use it directly.
func New(cfg config.Config) *slog.Logger {
	level := parseLevel(cfg.LOG_LEVEL)

	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(cfg.LOG_FORMAT)) {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	default:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// parseLevel maps a string level (debug/info/warn/error) to a slog.Level.
// Unknown values default to info.
func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithLogger returns a child logger with the given key-value attributes.
// Useful for request-scoped logging.
func WithLogger(logger *slog.Logger, attrs ...any) *slog.Logger {
	if logger == nil {
		logger = slog.Default()
	}
	return logger.With(attrs...)
}
