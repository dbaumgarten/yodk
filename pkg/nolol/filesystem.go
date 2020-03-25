package nolol

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// FileSystem defines an interface to get the content of files by name
// used so the converter can query for the content of files mentioned in include-directives
type FileSystem interface {
	Get(name string) (string, error)
}

// DiskFileSystem retrieves files from a directory on the disk
type DiskFileSystem struct {
	Dir string
}

// Get implements FileSystem
func (f DiskFileSystem) Get(name string) (string, error) {
	path := filepath.Join(f.Dir, name)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// MemoryFileSystem serves files from Memory
type MemoryFileSystem map[string]string

// Get implements FileSystem
func (f MemoryFileSystem) Get(name string) (string, error) {
	file, exists := f[name]
	if !exists {
		return "", fmt.Errorf("File not found")
	}
	return file, nil
}
