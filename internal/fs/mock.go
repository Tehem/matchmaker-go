package fs

import (
	"os"
	"path/filepath"
	"strings"
)

// MockFileSystem is a mock implementation of the filesystem interface for testing
type MockFileSystem struct {
	files       map[string][]byte
	directories map[string]struct{}
}

// NewMockFileSystem creates a new mock filesystem
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files:       make(map[string][]byte),
		directories: make(map[string]struct{}),
	}
}

// ReadFile reads a file from the mock filesystem
func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	path = filepath.Clean(path)
	if data, ok := m.files[path]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

// WriteFile writes a file to the mock filesystem
func (m *MockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	path = filepath.Clean(path)
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		if err := m.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	m.files[path] = data
	return nil
}

// MkdirAll creates a directory and all necessary parent directories
func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	path = filepath.Clean(path)
	// Split path into components
	parts := strings.Split(path, string(os.PathSeparator))
	current := ""

	// Create each directory in the path
	for _, part := range parts {
		if part == "" {
			continue
		}
		current = filepath.Join(current, part)
		m.directories[current] = struct{}{}
	}
	return nil
}
