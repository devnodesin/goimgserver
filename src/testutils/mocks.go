package testutils

import (
	"github.com/stretchr/testify/mock"
)

// MockImageProcessor is a mock implementation of an image processor
type MockImageProcessor struct {
	mock.Mock
}

// Process processes an image with the given dimensions and format
func (m *MockImageProcessor) Process(data []byte, width, height int, format string) ([]byte, error) {
	args := m.Called(data, width, height, format)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// Resize resizes an image to the given dimensions
func (m *MockImageProcessor) Resize(data []byte, width, height int) ([]byte, error) {
	args := m.Called(data, width, height)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// Convert converts an image to a different format
func (m *MockImageProcessor) Convert(data []byte, format string) ([]byte, error) {
	args := m.Called(data, format)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// MockCache is a mock implementation of a cache
type MockCache struct {
	mock.Mock
}

// Get retrieves a value from the cache
func (m *MockCache) Get(key string) ([]byte, error) {
	args := m.Called(key)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// Set stores a value in the cache
func (m *MockCache) Set(key string, value []byte) error {
	args := m.Called(key, value)
	return args.Error(0)
}

// Clear clears the cache
func (m *MockCache) Clear() error {
	args := m.Called()
	return args.Error(0)
}

// Delete removes a specific key from the cache
func (m *MockCache) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

// MockFileResolver is a mock implementation of a file resolver
type MockFileResolver struct {
	mock.Mock
}

// Resolve resolves a filename to a full path
func (m *MockFileResolver) Resolve(filename string) (string, error) {
	args := m.Called(filename)
	return args.String(0), args.Error(1)
}

// ResolveGroup resolves a grouped filename
func (m *MockFileResolver) ResolveGroup(group, filename string) (string, error) {
	args := m.Called(group, filename)
	return args.String(0), args.Error(1)
}

// MockGitManager is a mock implementation of a git manager
type MockGitManager struct {
	mock.Mock
}

// Update performs a git pull/update
func (m *MockGitManager) Update() error {
	args := m.Called()
	return args.Error(0)
}

// IsRepo checks if the directory is a git repository
func (m *MockGitManager) IsRepo() bool {
	args := m.Called()
	return args.Bool(0)
}

// GetStatus returns the git status
func (m *MockGitManager) GetStatus() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// MockLogger is a mock implementation of a logger
type MockLogger struct {
	mock.Mock
}

// Info logs an info message
func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

// Error logs an error message
func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

// Warn logs a warning message
func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.Called(msg, args)
}

// Debug logs a debug message
func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.Called(msg, args)
}
