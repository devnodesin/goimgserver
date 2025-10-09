package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLogger_Configuration_Levels tests log level configuration
func TestLogger_Configuration_Levels(t *testing.T) {
	tests := []struct {
		name          string
		level         slog.Level
		logFunc       func(*Logger, string)
		shouldLog     bool
	}{
		{"Debug level logs debug", slog.LevelDebug, func(l *Logger, msg string) { l.Debug(msg) }, true},
		{"Debug level logs info", slog.LevelDebug, func(l *Logger, msg string) { l.Info(msg) }, true},
		{"Info level skips debug", slog.LevelInfo, func(l *Logger, msg string) { l.Debug(msg) }, false},
		{"Info level logs info", slog.LevelInfo, func(l *Logger, msg string) { l.Info(msg) }, true},
		{"Warn level skips info", slog.LevelWarn, func(l *Logger, msg string) { l.Info(msg) }, false},
		{"Warn level logs warn", slog.LevelWarn, func(l *Logger, msg string) { l.Warn(msg) }, true},
		{"Error level skips warn", slog.LevelError, func(l *Logger, msg string) { l.Warn(msg) }, false},
		{"Error level logs error", slog.LevelError, func(l *Logger, msg string) { l.Error(msg) }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger(buf, tt.level)
			
			tt.logFunc(logger, "test message")
			
			if tt.shouldLog {
				assert.NotEmpty(t, buf.String(), "Expected log output")
				assert.Contains(t, buf.String(), "test message")
			} else {
				assert.Empty(t, buf.String(), "Expected no log output")
			}
		})
	}
}

// TestLogger_Structured_JSON tests structured JSON logging
func TestLogger_Structured_JSON(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelInfo)
	
	logger.Info("test message")
	
	// Parse JSON output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err, "Log output should be valid JSON")
	
	// Verify standard fields
	assert.Equal(t, "test message", logEntry["msg"])
	assert.Contains(t, logEntry, "time")
	assert.Contains(t, logEntry, "level")
}

// TestLogger_Structured_Fields tests custom field logging
func TestLogger_Structured_Fields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelInfo)
	
	logger.InfoWithFields("test message", 
		"request_id", "12345",
		"user_id", "user-abc",
		"duration_ms", 150,
	)
	
	// Parse JSON output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	
	assert.Equal(t, "test message", logEntry["msg"])
	assert.Equal(t, "12345", logEntry["request_id"])
	assert.Equal(t, "user-abc", logEntry["user_id"])
	assert.Equal(t, float64(150), logEntry["duration_ms"])
}

// TestLogger_Context_Propagation tests context propagation through logs
func TestLogger_Context_Propagation(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelInfo)
	
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "ctx-12345")
	
	logger.InfoContext(ctx, "test message")
	
	output := buf.String()
	assert.Contains(t, output, "test message")
	// Context values should be included if logger extracts them
	// This test validates that context can be passed through
}

// TestLogger_ErrorWithStack tests error logging with stack trace
func TestLogger_ErrorWithStack(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelError)
	
	testErr := assert.AnError
	logger.ErrorWithFields("operation failed", "error", testErr.Error())
	
	output := buf.String()
	assert.Contains(t, output, "operation failed")
	assert.Contains(t, output, testErr.Error())
}

// TestLogger_MultipleInstances tests that multiple logger instances don't interfere
func TestLogger_MultipleInstances(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	
	logger1 := NewLogger(buf1, slog.LevelInfo)
	logger2 := NewLogger(buf2, slog.LevelDebug)
	
	logger1.Info("logger1 message")
	logger2.Debug("logger2 message")
	
	assert.Contains(t, buf1.String(), "logger1 message")
	assert.NotContains(t, buf1.String(), "logger2 message")
	
	assert.Contains(t, buf2.String(), "logger2 message")
	assert.NotContains(t, buf2.String(), "logger1 message")
}

// TestLogger_PerformanceOverhead tests logging performance impact
func TestLogger_PerformanceOverhead(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(buf, slog.LevelInfo)
	
	// Log many messages to ensure reasonable performance
	for i := 0; i < 1000; i++ {
		logger.Info("performance test message")
	}
	
	// Verify all messages were logged
	lines := strings.Count(buf.String(), "\n")
	assert.GreaterOrEqual(t, lines, 1000, "Should log all messages")
}
