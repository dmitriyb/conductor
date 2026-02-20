package config

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"
)

// TestFR5_InitLoggingReturnsLogger verifies that InitLogging returns a
// non-nil *slog.Logger for valid level strings.
func TestFR5_InitLoggingReturnsLogger(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error", "INFO", "Debug", "unknown", ""}
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			logger := InitLogging(level, io.Discard)
			if logger == nil {
				t.Fatal("InitLogging returned nil")
			}
		})
	}
}

// TestFR5_LevelParsing verifies that the returned logger respects the
// configured level by checking which messages are enabled.
func TestFR5_LevelParsing(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantLevel slog.Level
	}{
		{"debug", "debug", slog.LevelDebug},
		{"info", "info", slog.LevelInfo},
		{"warn", "warn", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"default for unknown", "bogus", slog.LevelInfo},
		{"default for empty", "", slog.LevelInfo},
		{"case insensitive", "DEBUG", slog.LevelDebug},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := InitLogging(tt.level, io.Discard)
			if !logger.Enabled(nil, tt.wantLevel) {
				t.Errorf("level %s should be enabled for input %q", tt.wantLevel, tt.level)
			}
			// A level below the configured one should be disabled
			// (except debug, which is the lowest).
			if tt.wantLevel > slog.LevelDebug {
				belowLevel := tt.wantLevel - 4 // slog levels are spaced by 4
				if logger.Enabled(nil, belowLevel) {
					t.Errorf("level %s should be disabled for input %q", belowLevel, tt.level)
				}
			}
		})
	}
}

// TestFR5_NonTTYUsesJSONHandler verifies that newHandler with isTTY=false
// produces JSON output.
func TestFR5_NonTTYUsesJSONHandler(t *testing.T) {
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	h := newHandler(&buf, false, opts)
	logger := slog.New(h)
	logger.Info("test message", "key", "value")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("expected JSON output, got: %s", buf.String())
	}
	if msg, ok := m["msg"].(string); !ok || msg != "test message" {
		t.Errorf("msg = %v, want %q", m["msg"], "test message")
	}
}

// TestFR5_TTYUsesTextHandler verifies that newHandler with isTTY=true
// produces non-JSON key=value output.
func TestFR5_TTYUsesTextHandler(t *testing.T) {
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	h := newHandler(&buf, true, opts)
	logger := slog.New(h)
	logger.Info("test message", "key", "value")

	output := buf.String()
	// Text handler output contains key=value pairs, not JSON braces.
	if strings.Contains(output, "{") {
		t.Errorf("text handler produced JSON-like output: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("expected key=value in text output, got: %s", output)
	}
}

// TestFR5_ComponentAttribute verifies that logger.With adds structured
// component attributes to log output.
func TestFR5_ComponentAttribute(t *testing.T) {
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	h := newHandler(&buf, false, opts)
	logger := slog.New(h)

	compLogger := logger.With("component", "agent")
	compLogger.Info("step completed", "step", "build")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("expected JSON output, got: %s", buf.String())
	}
	if comp, ok := m["component"].(string); !ok || comp != "agent" {
		t.Errorf("component = %v, want %q", m["component"], "agent")
	}
	if step, ok := m["step"].(string); !ok || step != "build" {
		t.Errorf("step = %v, want %q", m["step"], "build")
	}
}

// TestFR5_NoGlobalLoggerState verifies that InitLogging does not set the
// default slog logger (D3 requirement).
func TestFR5_NoGlobalLoggerState(t *testing.T) {
	defaultBefore := slog.Default()
	_ = InitLogging("info", io.Discard)
	defaultAfter := slog.Default()
	if defaultBefore != defaultAfter {
		t.Error("InitLogging must not call slog.SetDefault")
	}
}
