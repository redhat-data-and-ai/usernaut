package logger

import (
	"fmt"
	"os"

	"golang.org/x/exp/slog"
)

// LoggerConfig defines the configuration for logging.
type LoggerConfig struct {
	Level   string
	Enabled bool
}

const (
	DEFAULT_LOG_ENABLED = false
)

type DiscardWriter struct{}

// Write implements the io.Writer interface.
func (d DiscardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil // Discard all data written
}

type Logger struct {
	Log *slog.Logger
}

var Log Logger

// NewLogger initializes a new logger with the provided configuration.
func NewLogger(cfg *LoggerConfig) Logger {
	loggingEnabled := DEFAULT_LOG_ENABLED
	if cfg != nil {
		loggingEnabled = cfg.Enabled
	}

	var logger *slog.Logger

	options := slog.HandlerOptions{}

	logLevel := cfg.Level
	switch logLevel {
	case "info":
		options.Level = slog.LevelInfo
	case "error":
		options.Level = slog.LevelError
	case "debug":
		options.Level = slog.LevelDebug
	default:
		options.Level = slog.LevelInfo
	}

	if loggingEnabled {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &options))
	} else {
		logger = slog.New(slog.NewJSONHandler(DiscardWriter{}, nil))
	}

	l := Logger{}
	l.Log = logger
	Log = l
	return l
}

// Debugf logs a formatted debug message.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Log.Debug(fmt.Sprintf(format, args...))
}

// Debug logs a debug message.
func (l *Logger) Debug(message string) {
	l.Log.Debug(message)
}

// Infof logs a formatted info message.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Log.Info(fmt.Sprintf(format, args...))
}

// Info logs an info message.
func (l *Logger) Info(message string) {
	l.Log.Info(message)
}

// Errorf logs a formatted error message.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Log.Error(fmt.Sprintf(format, args...))
}

// Error logs an error message.
func (l *Logger) Error(message string) {
	l.Log.Error(message)
}
