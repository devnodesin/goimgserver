package logging

import (
	"errors"
	"log/slog"
)

// Config represents logging configuration
type Config struct {
	Level      slog.Level
	JSONFormat bool
	FilePath   string
	MaxSize    int64
	MaxBackups int
	AddSource  bool
}

// DefaultConfig returns default logging configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      slog.LevelInfo,
		JSONFormat: false,
		FilePath:   "",
		MaxSize:    100 * 1024 * 1024, // 100MB
		MaxBackups: 5,
		AddSource:  false,
	}
}

// ProductionConfig returns production-optimized configuration
func ProductionConfig() *Config {
	return &Config{
		Level:      slog.LevelInfo,
		JSONFormat: true,
		FilePath:   "",
		MaxSize:    100 * 1024 * 1024,
		MaxBackups: 10,
		AddSource:  false,
	}
}

// DevelopmentConfig returns development-optimized configuration
func DevelopmentConfig() *Config {
	return &Config{
		Level:      slog.LevelDebug,
		JSONFormat: false,
		FilePath:   "",
		MaxSize:    50 * 1024 * 1024,
		MaxBackups: 3,
		AddSource:  true,
	}
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
	if c.MaxSize <= 0 {
		return errors.New("max size must be greater than 0")
	}
	if c.MaxBackups < 0 {
		return errors.New("max backups cannot be negative")
	}
	return nil
}

// Clone creates a deep copy of the configuration
func (c *Config) Clone() *Config {
	return &Config{
		Level:      c.Level,
		JSONFormat: c.JSONFormat,
		FilePath:   c.FilePath,
		MaxSize:    c.MaxSize,
		MaxBackups: c.MaxBackups,
		AddSource:  c.AddSource,
	}
}
