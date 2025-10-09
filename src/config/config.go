package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config holds all application configuration
type Config struct {
	Port             int
	ImagesDir        string
	CacheDir         string
	Dump             bool
	DefaultImagePath string
	PreCacheEnabled  bool
	PreCacheWorkers  int
}

// ParseArgs parses command-line arguments and returns a Config
func ParseArgs(args []string) (*Config, error) {
	fs := flag.NewFlagSet("goimgserver", flag.ContinueOnError)

	cfg := &Config{}

	fs.IntVar(&cfg.Port, "port", 9000, "Server port")
	fs.StringVar(&cfg.ImagesDir, "imagesdir", "./images", "Images directory")
	fs.StringVar(&cfg.CacheDir, "cachedir", "./cache", "Cache directory")
	fs.BoolVar(&cfg.Dump, "dump", false, "Dump settings to settings.conf")
	fs.BoolVar(&cfg.PreCacheEnabled, "precache", true, "Enable pre-caching of images on startup")
	fs.IntVar(&cfg.PreCacheWorkers, "precache-workers", 0, "Number of workers for pre-cache (0 = auto, uses CPU count)")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration and creates directories if needed
func (c *Config) Validate() error {
	// Validate port range
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}

	// Ensure directories exist, create if missing
	if err := os.MkdirAll(c.ImagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create images directory: %w", err)
	}

	if err := os.MkdirAll(c.CacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	return nil
}

// DumpSettings writes current configuration to a file
func (c *Config) DumpSettings(filename string) error {
	content := c.String()
	return os.WriteFile(filename, []byte(content), 0644)
}

// String returns a string representation of the configuration
func (c *Config) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Port: %d\n", c.Port))
	sb.WriteString(fmt.Sprintf("ImagesDir: %s\n", c.ImagesDir))
	sb.WriteString(fmt.Sprintf("CacheDir: %s\n", c.CacheDir))
	sb.WriteString(fmt.Sprintf("Dump: %v\n", c.Dump))
	if c.DefaultImagePath != "" {
		sb.WriteString(fmt.Sprintf("DefaultImagePath: %s\n", c.DefaultImagePath))
	}
	sb.WriteString(fmt.Sprintf("PreCacheEnabled: %v\n", c.PreCacheEnabled))
	sb.WriteString(fmt.Sprintf("PreCacheWorkers: %d\n", c.PreCacheWorkers))
	return sb.String()
}
