package config

import (
	"log/slog"
	"os"
	"strings"

	"golang.org/x/term"
)

// InitLogging creates a *slog.Logger configured for the given level string.
// It uses a text handler when stderr is a TTY, and a JSON handler otherwise.
// Supported levels: debug, info, warn, error. Unknown levels default to info.
func InitLogging(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: lvl}
	var h slog.Handler
	if term.IsTerminal(int(os.Stderr.Fd())) {
		h = slog.NewTextHandler(os.Stderr, opts)
	} else {
		h = slog.NewJSONHandler(os.Stderr, opts)
	}
	return slog.New(h)
}
