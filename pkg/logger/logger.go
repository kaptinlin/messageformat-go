// Package logger provides simple structured logging for MessageFormat 2.0
package logger

import (
	"io"
	"log/slog"
	"os"
)

// Global logger instance
var global *slog.Logger

func init() {
	// Default configuration: INFO level, text format, output to stderr
	global = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// SetLogger allows users to completely override the global logger
func SetLogger(logger *slog.Logger) {
	global = logger
}

// GetLogger returns the current global logger
func GetLogger() *slog.Logger {
	return global
}

// SetLevel sets the global logging level
func SetLevel(level slog.Level) {
	global = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
}

// SetJSON switches the global logger to JSON format
func SetJSON() {
	global = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// SetOutput sets the output destination for the global logger
func SetOutput(w io.Writer) {
	global = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// Global logging methods
func Debug(msg string, args ...any) {
	global.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	global.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	global.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	global.Error(msg, args...)
}
