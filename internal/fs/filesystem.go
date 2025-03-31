package fs

import (
	"os"
)

// FileSystem defines the interface for file operations
type FileSystem interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
}

// DefaultFileSystem implements FileSystem using the real filesystem
type DefaultFileSystem struct{}

func (DefaultFileSystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (DefaultFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (DefaultFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Default is the default filesystem implementation
var Default FileSystem = DefaultFileSystem{}
