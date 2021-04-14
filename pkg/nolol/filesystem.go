package nolol

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/dbaumgarten/yodk/stdlib"
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

func addExtension(imp string) string {
	if !strings.HasSuffix(imp, ".nolol") {
		return imp + ".nolol"
	}
	return imp
}

func fixPath(name string) string {
	return filepath.FromSlash(name)
}

// Get implements FileSystem
func (f DiskFileSystem) Get(name string) (string, error) {
	name = addExtension(name)

	if stdlib.Is(name) {
		file, err := stdlib.Get(name)
		if err != nil {
			return "", err
		}
		return file, nil
	}

	name = fixPath(name)
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
	name = addExtension(name)
	if stdlib.Is(name) {
		file, err := stdlib.Get(name)
		if err != nil {
			return "", err
		}
		return file, nil
	}

	file, exists := f[name]
	if !exists {
		return "", fmt.Errorf("File not found")
	}
	return file, nil
}
