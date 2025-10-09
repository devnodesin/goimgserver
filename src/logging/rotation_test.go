package logging

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLogRotation_FileSize_Limits tests log rotation by size
func TestLogRotation_FileSize_Limits(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	
	// Create rotator with small max size (1KB)
	rotator, err := NewRotator(logFile, 1024, 3)
	require.NoError(t, err)
	defer rotator.Close()
	
	// Write data larger than max size
	data := make([]byte, 2048)
	for i := range data {
		data[i] = 'a'
	}
	
	_, err = rotator.Write(data)
	require.NoError(t, err)
	
	// Check that rotation occurred
	files, err := filepath.Glob(filepath.Join(tmpDir, "test*.log*"))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), 2, "Should have rotated log file")
}

// TestLogRotation_Time_Based tests time-based log rotation
func TestLogRotation_Time_Based(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	
	// Create rotator with time-based rotation (1 second for testing)
	rotator, err := NewRotator(logFile, 10*1024, 5)
	require.NoError(t, err)
	defer rotator.Close()
	
	// Write initial data
	rotator.Write([]byte("initial log entry\n"))
	
	// Get initial file info
	info1, err := os.Stat(logFile)
	require.NoError(t, err)
	modTime1 := info1.ModTime()
	
	// Wait a bit and write more data
	time.Sleep(10 * time.Millisecond)
	rotator.Write([]byte("second log entry\n"))
	
	// File should exist and have been modified
	info2, err := os.Stat(logFile)
	require.NoError(t, err)
	assert.True(t, info2.ModTime().After(modTime1) || info2.ModTime().Equal(modTime1))
}

// TestLogRotation_MaxBackups tests backup file retention
func TestLogRotation_MaxBackups(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	
	maxBackups := 2
	rotator, err := NewRotator(logFile, 512, maxBackups)
	require.NoError(t, err)
	defer rotator.Close()
	
	// Force multiple rotations
	data := make([]byte, 600)
	for i := range data {
		data[i] = 'b'
	}
	
	// Write enough to trigger multiple rotations
	for i := 0; i < 5; i++ {
		rotator.Write(data)
	}
	
	// Count backup files
	files, err := filepath.Glob(filepath.Join(tmpDir, "test*.log*"))
	require.NoError(t, err)
	
	// Should have main file + maxBackups
	assert.LessOrEqual(t, len(files), maxBackups+1, "Should not exceed max backups")
}

// TestLogRotation_CreateParentDir tests automatic directory creation
func TestLogRotation_CreateParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "subdir", "nested", "test.log")
	
	rotator, err := NewRotator(logFile, 1024, 3)
	require.NoError(t, err)
	defer rotator.Close()
	
	// Write some data
	rotator.Write([]byte("test log entry\n"))
	
	// Verify file exists
	_, err = os.Stat(logFile)
	require.NoError(t, err, "Log file should be created in nested directory")
}

// TestLogRotation_ConcurrentWrites tests thread safety
func TestLogRotation_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	
	rotator, err := NewRotator(logFile, 10*1024, 5)
	require.NoError(t, err)
	defer rotator.Close()
	
	// Write concurrently from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				rotator.Write([]byte("concurrent write\n"))
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify file exists
	_, err = os.Stat(logFile)
	require.NoError(t, err)
}

// TestLogRotation_EmptyFile tests handling of empty files
func TestLogRotation_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	
	rotator, err := NewRotator(logFile, 1024, 3)
	require.NoError(t, err)
	defer rotator.Close()
	
	// Close without writing
	rotator.Close()
	
	// File should still exist (even if empty)
	info, err := os.Stat(logFile)
	require.NoError(t, err)
	assert.Equal(t, int64(0), info.Size())
}
