package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Rotator handles log file rotation based on size and backups
type Rotator struct {
	mu         sync.Mutex
	file       *os.File
	filename   string
	maxSize    int64
	maxBackups int
	size       int64
}

// NewRotator creates a new log rotator
func NewRotator(filename string, maxSize int64, maxBackups int) (*Rotator, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	
	r := &Rotator{
		filename:   filename,
		maxSize:    maxSize,
		maxBackups: maxBackups,
	}
	
	// Open or create the log file
	if err := r.openFile(); err != nil {
		return nil, err
	}
	
	return r, nil
}

// Write implements io.Writer interface
func (r *Rotator) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check if rotation is needed
	if r.size+int64(len(p)) > r.maxSize {
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}
	
	// Write to current file
	n, err = r.file.Write(p)
	r.size += int64(n)
	return n, err
}

// Close closes the log file
func (r *Rotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// openFile opens the log file and gets its current size
func (r *Rotator) openFile() error {
	file, err := os.OpenFile(r.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}
	
	r.file = file
	r.size = info.Size()
	return nil
}

// rotate performs log file rotation
func (r *Rotator) rotate() error {
	// Close current file
	if r.file != nil {
		r.file.Close()
	}
	
	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s.%s", r.filename, timestamp)
	
	// Rename current file to backup
	if err := os.Rename(r.filename, backupName); err != nil && !os.IsNotExist(err) {
		return err
	}
	
	// Clean up old backups
	r.cleanupOldBackups()
	
	// Open new file
	return r.openFile()
}

// cleanupOldBackups removes old backup files exceeding maxBackups
func (r *Rotator) cleanupOldBackups() {
	// Find all backup files
	pattern := r.filename + ".*"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	
	// Sort by modification time (oldest first)
	sort.Slice(matches, func(i, j int) bool {
		info1, err1 := os.Stat(matches[i])
		info2, err2 := os.Stat(matches[j])
		if err1 != nil || err2 != nil {
			return false
		}
		return info1.ModTime().Before(info2.ModTime())
	})
	
	// Remove oldest files if exceeding maxBackups
	if len(matches) > r.maxBackups {
		for _, file := range matches[:len(matches)-r.maxBackups] {
			os.Remove(file)
		}
	}
}

// Ensure Rotator implements io.WriteCloser
var _ io.WriteCloser = (*Rotator)(nil)
