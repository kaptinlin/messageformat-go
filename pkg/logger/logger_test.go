package logger

import (
	"bytes"
	"log/slog"
	"testing"
)

func TestBasicLogging(t *testing.T) {
	// Test basic logging functionality
	Debug("debug message", "key", "value")
	Info("info message", "key", "value")
	Warn("warn message", "key", "value")
	Error("error message", "key", "value")
}

func TestSetLevel(t *testing.T) {
	SetLevel(slog.LevelDebug)
	Debug("this should appear")

	SetLevel(slog.LevelError)
	Debug("this should not appear")
}

func TestSetLogger(t *testing.T) {
	var buf bytes.Buffer
	customLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	SetLogger(customLogger)
	Info("global test message")

	if !bytes.Contains(buf.Bytes(), []byte("global test message")) {
		t.Error("Expected global logger to be replaced")
	}
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	if logger == nil {
		t.Error("Expected logger to be available")
	}
}
