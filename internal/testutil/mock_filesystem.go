package testutil

import (
	"fmt"
	"os"
)

// MockFileSystem implements a simple in-memory file system for testing
type MockFileSystem struct {
	files map[string][]byte
}

// NewMockFileSystem creates a new mock filesystem
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
	}
}

// ReadFile implements FileSystem
func (m *MockFileSystem) ReadFile(filename string) ([]byte, error) {
	if data, ok := m.files[filename]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("file not found: %s", filename)
}

// WriteFile implements FileSystem
func (m *MockFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	m.files[filename] = data
	return nil
}
