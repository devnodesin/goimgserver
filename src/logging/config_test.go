package logging

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfig_DefaultValues tests default configuration values
func TestConfig_DefaultValues(t *testing.T) {
	config := DefaultConfig()
	
	assert.Equal(t, slog.LevelInfo, config.Level)
	assert.False(t, config.JSONFormat)
	assert.Equal(t, "", config.FilePath)
	assert.Equal(t, int64(100*1024*1024), config.MaxSize) // 100MB
	assert.Equal(t, 5, config.MaxBackups)
	assert.False(t, config.AddSource)
}

// TestConfig_ProductionDefaults tests production configuration
func TestConfig_ProductionDefaults(t *testing.T) {
	config := ProductionConfig()
	
	assert.Equal(t, slog.LevelInfo, config.Level)
	assert.True(t, config.JSONFormat)
	assert.False(t, config.AddSource)
}

// TestConfig_DevelopmentDefaults tests development configuration
func TestConfig_DevelopmentDefaults(t *testing.T) {
	config := DevelopmentConfig()
	
	assert.Equal(t, slog.LevelDebug, config.Level)
	assert.False(t, config.JSONFormat)
	assert.True(t, config.AddSource)
}

// TestConfig_CustomLevel tests custom log level configuration
func TestConfig_CustomLevel(t *testing.T) {
	tests := []struct {
		name  string
		level slog.Level
	}{
		{"Debug", slog.LevelDebug},
		{"Info", slog.LevelInfo},
		{"Warn", slog.LevelWarn},
		{"Error", slog.LevelError},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Level = tt.level
			
			assert.Equal(t, tt.level, config.Level)
		})
	}
}

// TestConfig_FileRotation tests file rotation configuration
func TestConfig_FileRotation(t *testing.T) {
	config := DefaultConfig()
	config.FilePath = "/var/log/app.log"
	config.MaxSize = 50 * 1024 * 1024  // 50MB
	config.MaxBackups = 10
	
	assert.Equal(t, "/var/log/app.log", config.FilePath)
	assert.Equal(t, int64(50*1024*1024), config.MaxSize)
	assert.Equal(t, 10, config.MaxBackups)
}

// TestConfig_Validation tests configuration validation
func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		shouldError bool
	}{
		{
			name: "Valid config",
			config: &Config{
				Level:      slog.LevelInfo,
				MaxSize:    1024,
				MaxBackups: 5,
			},
			shouldError: false,
		},
		{
			name: "Zero max size",
			config: &Config{
				Level:      slog.LevelInfo,
				MaxSize:    0,
				MaxBackups: 5,
			},
			shouldError: true,
		},
		{
			name: "Negative max backups",
			config: &Config{
				Level:      slog.LevelInfo,
				MaxSize:    1024,
				MaxBackups: -1,
			},
			shouldError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestConfig_Clone tests configuration cloning
func TestConfig_Clone(t *testing.T) {
	original := &Config{
		Level:      slog.LevelDebug,
		JSONFormat: true,
		FilePath:   "/var/log/test.log",
		MaxSize:    2048,
		MaxBackups: 3,
		AddSource:  true,
	}
	
	cloned := original.Clone()
	
	// Verify values are equal
	assert.Equal(t, original.Level, cloned.Level)
	assert.Equal(t, original.JSONFormat, cloned.JSONFormat)
	assert.Equal(t, original.FilePath, cloned.FilePath)
	assert.Equal(t, original.MaxSize, cloned.MaxSize)
	assert.Equal(t, original.MaxBackups, cloned.MaxBackups)
	assert.Equal(t, original.AddSource, cloned.AddSource)
	
	// Modify clone shouldn't affect original
	cloned.Level = slog.LevelError
	assert.NotEqual(t, original.Level, cloned.Level)
}
