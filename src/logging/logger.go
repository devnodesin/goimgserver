package logging

import (
	"context"
	"io"
	"log/slog"
)

// Logger wraps slog.Logger with convenience methods
type Logger struct {
	logger *slog.Logger
}

// NewLogger creates a new logger with specified output and level
func NewLogger(w io.Writer, level slog.Level) *Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewJSONHandler(w, opts)
	return &Logger{
		logger: slog.New(handler),
	}
}

// NewLoggerFromConfig creates a logger from configuration
func NewLoggerFromConfig(w io.Writer, config *Config) *Logger {
	opts := &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.AddSource,
	}
	
	var handler slog.Handler
	if config.JSONFormat {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}
	
	return &Logger{
		logger: slog.New(handler),
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug(msg)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info(msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn(msg)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.logger.Error(msg)
}

// InfoWithFields logs an info message with structured fields
func (l *Logger) InfoWithFields(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// ErrorWithFields logs an error message with structured fields
func (l *Logger) ErrorWithFields(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

// WarnWithFields logs a warning message with structured fields
func (l *Logger) WarnWithFields(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// DebugWithFields logs a debug message with structured fields
func (l *Logger) DebugWithFields(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// InfoContext logs an info message with context
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.InfoContext(ctx, msg, args...)
}

// ErrorContext logs an error message with context
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.ErrorContext(ctx, msg, args...)
}

// WarnContext logs a warning message with context
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.WarnContext(ctx, msg, args...)
}

// DebugContext logs a debug message with context
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.DebugContext(ctx, msg, args...)
}

// With returns a new Logger with additional fields
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{
		logger: l.logger.With(args...),
	}
}
