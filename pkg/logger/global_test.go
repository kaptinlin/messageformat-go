package logger

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert.Contains(t, buf.String(), "global test message", "Expected global logger to be replaced")
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger()
	assert.NotNil(t, logger, "Expected logger to be available")
}

func TestSetJSON(t *testing.T) {
	var buf bytes.Buffer
	original := GetLogger()
	t.Cleanup(func() { SetLogger(original) })

	SetOutput(&buf)
	Info("text message")
	assert.Contains(t, buf.String(), "msg=\"text message\"")

	jsonLogger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	buf.Reset()
	SetJSON()
	SetLogger(jsonLogger)
	Info("json message", "key", "value")

	out := buf.String()
	assert.Contains(t, out, `"msg":"json message"`)
	assert.Contains(t, out, `"key":"value"`)
}

func TestSetOutput(t *testing.T) {
	var buf bytes.Buffer
	original := GetLogger()
	t.Cleanup(func() { SetLogger(original) })

	SetOutput(&buf)
	Info("custom output", "answer", 42)

	out := buf.String()
	assert.Contains(t, out, "custom output")
	assert.Contains(t, out, "answer=42")
}
