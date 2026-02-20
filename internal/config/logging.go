package config

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"golang.org/x/term"
)

// InitLogging creates a *slog.Logger configured for the given level string.
// It uses a text handler when w is a TTY, and a JSON handler otherwise.
// Supported levels: debug, info, warn, error. Unknown levels default to info.
func InitLogging(level string, w io.Writer) *slog.Logger {
	lvl := parseLevel(level)
	opts := &slog.HandlerOptions{Level: lvl}
	isTTY := false
	if f, ok := w.(*os.File); ok {
		isTTY = term.IsTerminal(int(f.Fd()))
	}
	h := newHandler(w, isTTY, opts)
	return slog.New(h)
}

// parseLevel maps a level string to a slog.Level.
// Unknown or empty strings default to slog.LevelInfo.
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// newHandler creates a slog.Handler that writes to w.
// It returns a TextHandler for TTY output, and a JSONHandler otherwise.
func newHandler(w io.Writer, isTTY bool, opts *slog.HandlerOptions) slog.Handler {
	if isTTY {
		return slog.NewTextHandler(w, opts)
	}
	return slog.NewJSONHandler(w, opts)
}
