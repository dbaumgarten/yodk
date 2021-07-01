// Code generated for package stdlib by go-bindata DO NOT EDIT. (@generated)
// sources:
// src/logic.nolol
// src/math_advanced.nolol
// src/math_basic.nolol
// src/math_professional.nolol
// src/string.nolol
package stdlib

import (
	"github.com/elazarl/go-bindata-assetfs"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _logicNolol = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x52\x4f\x6f\xd3\x4e\x10\xbd\xfb\x53\xbc\x5f\x54\xa9\xc9\x8f\x04\x37\x1c\x29\xe1\xc0\xad\x57\x04\xe7\x6a\xd7\x1e\xdb\xa3\xda\x33\xd6\xec\x2c\x29\xdf\x1e\xad\x43\xd3\xa4\xe5\x64\xc9\xf3\xf4\xfe\x6e\x5d\xe3\xc7\xc0\x09\x1d\x8f\x84\x46\xc5\x03\x4b\x42\x0c\x89\x1b\xb4\xd4\xb1\xb0\xb3\x4a\x42\x90\x16\x53\x68\x4c\x13\x3a\x35\x8c\xda\x73\xb3\xd3\x99\x2c\x2c\xf7\xaa\xae\xf1\x30\xcd\x6a\x8e\x9c\x58\x7a\xdc\x82\xa5\x19\x73\x4b\x58\x25\x6f\xeb\x05\xbf\xc2\x6d\x55\x80\xdf\xc9\xb3\x15\x4e\x70\x57\x34\xdb\x45\x03\x9c\xe0\x96\x69\x0b\xf5\x81\xec\xc8\x89\x10\x0b\xfc\x8c\xd8\x22\x2c\x3e\x22\xa6\x9c\x1c\x91\x20\x79\x8a\x64\xe9\xe3\x05\xcb\xcb\xe9\x0e\x6a\xd8\x57\x8b\xe7\x93\xdd\x47\x27\x93\x60\xbf\xd7\x97\x84\x5b\xc4\x0d\xe8\x79\xb6\x0a\x00\x22\x3e\x60\x1d\x76\x71\xf3\xff\x19\x53\x91\xb4\x8b\xeb\x87\x37\x5e\xef\xb6\x98\x4d\xdb\xdc\x50\x49\x62\x59\x9c\x27\xda\x91\x99\x1a\x7c\x08\x8e\x23\x8f\x23\xd2\x13\xcf\xf0\x81\x60\x34\x05\x96\x52\xcd\xc8\x42\x85\xf0\x67\x0a\x3d\x7d\xc6\xea\x64\xae\x74\xcf\x92\xe9\xb1\x9c\xd7\xbf\x82\x6d\xee\xd1\xea\x61\x7f\x8f\xe4\xb9\xeb\x0e\x9f\x70\xb3\xaa\x96\xb5\x08\x37\x45\x9f\x97\xba\x83\x38\x5c\x21\xea\x7f\xa5\x54\x31\xe5\x66\xb8\x0a\x7e\xcd\x7d\x4e\xb1\xf9\xc2\xbd\xa8\xd1\xd7\x93\xa5\x52\xc0\xe9\x47\x7d\x78\x1f\xff\x65\xb4\x7d\x19\x2d\x94\x72\xe3\xeb\x62\x31\xfb\x62\x21\xaa\x0f\x05\xfc\x76\xa7\x7f\x8c\xf1\xac\xb6\x0e\xdb\xcb\xf2\x03\xfe\x3b\x20\x9e\x05\xbf\x8d\xda\x3c\x25\x84\x84\x51\xa5\x2f\xdf\x77\x4f\xe5\x8a\xf0\x18\xd8\x2f\xb2\xbd\x66\x1a\xa8\x44\xbc\x7a\x6a\x3e\x90\xa0\x57\xd7\xe5\x88\x22\x49\xd2\xfe\x09\x00\x00\xff\xff\x58\x96\xdd\x23\x0c\x03\x00\x00")

func logicNololBytes() ([]byte, error) {
	return bindataRead(
		_logicNolol,
		"logic.nolol",
	)
}

func logicNolol() (*asset, error) {
	bytes, err := logicNololBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "logic.nolol", size: 780, mode: os.FileMode(438), modTime: time.Unix(1623961616, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _math_advancedNolol = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x64\xcd\xc1\x4a\x04\x31\x0c\xc6\xf1\xb3\x7d\x8a\x8f\x01\x59\x3d\x68\xf1\x31\xbc\x8a\xf7\xa5\x33\x4d\x77\x02\x9d\x64\x69\x52\xed\xe3\xcb\xd6\x83\xc2\x1e\x13\x3e\x7e\xff\x18\xf1\xb9\xb3\xa1\x70\x25\x6c\x2a\x9e\x58\x0c\x6b\x32\xde\x90\xa9\xb0\xb0\xb3\x8a\x21\x49\xc6\x91\xb6\xa6\x86\xa2\x0d\x47\xf2\x3d\xc4\x88\xf7\xe3\xaa\xcd\xd1\x8d\xe5\x82\x13\x58\xb6\xda\x33\x61\x31\xcf\xf1\xb6\x59\x70\x0a\xe1\xee\x7b\x9e\xfe\x12\x6e\xc2\x07\x79\x6f\x62\xf0\x9d\x20\x34\x1c\x55\xbf\xa9\x81\xc5\xe9\x42\x0d\xae\x18\x61\x86\x67\xf3\x5c\xaa\x6a\x7b\x1a\xcf\xa0\x71\x6d\xe1\x61\xbc\x8c\xc7\xb7\x40\x92\xef\xac\xb4\x9a\xd6\xee\x84\xaf\x54\x3b\x41\x0b\xc6\xeb\x7f\x28\xad\xf6\xc7\xfc\x1e\xd3\xf9\x09\x00\x00\xff\xff\x04\xf6\x9d\xc7\x11\x01\x00\x00")

func math_advancedNololBytes() ([]byte, error) {
	return bindataRead(
		_math_advancedNolol,
		"math_advanced.nolol",
	)
}

func math_advancedNolol() (*asset, error) {
	bytes, err := math_advancedNololBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "math_advanced.nolol", size: 273, mode: os.FileMode(438), modTime: time.Unix(1625163899, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _math_basicNolol = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x84\x8f\x4f\x6f\xe2\x30\x10\x47\xcf\x9b\x4f\xf1\x13\x17\xfe\x08\x88\xb3\xbb\xd2\xee\x81\x70\xe7\x5a\xf5\x5e\x99\x64\x4c\x46\x72\xec\xc8\x33\x29\xfe\xf8\x55\x02\xa2\xb4\x97\x5e\x6c\x59\x9e\xf7\x9e\x5d\x96\x78\xed\x58\xe0\xd8\x13\x9a\x18\xd4\x72\x10\x9c\xad\x70\x83\x96\x1c\x07\x56\x8e\x41\x60\x43\x8b\xde\x36\x29\x0a\x5c\x4c\xe8\xad\x76\x45\x59\xe2\xd4\x0f\x31\x29\x46\xe1\x70\xc1\x12\x1c\x1a\x3f\xb6\x84\x85\x68\x5b\x4e\x33\x0b\x2c\x8b\x62\x6e\xd0\xcc\x50\x6f\x95\x1b\xeb\xa7\x94\xa8\x0d\x8a\x81\x8b\xb9\x73\xbb\x7f\x1b\xb8\xfe\xb3\xaf\xfe\x56\x3f\x50\xf4\x05\xa2\xfa\xf7\xfe\x5f\xf5\x7f\x66\x5e\x48\xc7\x14\x04\xda\x11\xec\x59\xa2\x1f\x95\xf0\x6e\xfd\x48\x88\x0e\x79\x5f\xcc\xbf\xb8\x71\xf6\x2c\xab\xbc\x06\xe5\x21\x15\xbf\x56\xf9\x58\x9b\xf5\x26\xef\x56\xf9\x30\xed\x05\x85\x76\x36\x9e\x1c\x32\x58\x70\x34\x5b\xa4\xbb\xbd\x02\x3b\x1c\xcc\xfd\x8c\x5d\xb5\x45\xd4\x8e\xd2\x95\x85\x1e\x43\xe6\xb9\x25\x7c\x09\xdf\x63\x53\xaa\x36\xeb\x47\xe9\xf9\xed\x81\xb2\xc2\xc7\x2b\x25\x70\x50\xba\x50\x82\x46\xe4\x67\xa5\xf3\x31\xa6\x4f\x67\x2e\x2b\x63\xcc\x66\x5a\x26\xe3\x47\x00\x00\x00\xff\xff\xee\x3a\x78\x8a\xda\x01\x00\x00")

func math_basicNololBytes() ([]byte, error) {
	return bindataRead(
		_math_basicNolol,
		"math_basic.nolol",
	)
}

func math_basicNolol() (*asset, error) {
	bytes, err := math_basicNololBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "math_basic.nolol", size: 474, mode: os.FileMode(438), modTime: time.Unix(1625166298, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _math_professionalNolol = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x64\xca\x31\x0e\xc2\x30\x0c\x05\xd0\x3d\xa7\xf8\xea\xd2\xd1\xe7\x60\x67\x47\xc6\x4e\xa8\xa5\xc6\x46\xb1\xcb\xf9\x11\x33\xeb\xd3\x23\xc2\xfd\xb0\xc4\xb0\xb3\x43\xc2\x8b\xcd\x13\x4f\x4e\x13\x68\x1f\xe6\x56\x16\x9e\x60\x57\x4c\x96\x15\x89\x11\x0b\x93\xeb\x68\x44\xb8\xcd\x77\xac\xc2\x95\xe6\x2f\xec\x30\x97\xf3\xd2\x8e\x2d\x4b\xe9\x77\x36\xec\xad\xfd\xe9\x83\xf5\xc3\x2e\x5d\xb7\xf6\x0d\x00\x00\xff\xff\xc8\xe5\x98\x34\x80\x00\x00\x00")

func math_professionalNololBytes() ([]byte, error) {
	return bindataRead(
		_math_professionalNolol,
		"math_professional.nolol",
	)
}

func math_professionalNolol() (*asset, error) {
	bytes, err := math_professionalNololBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "math_professional.nolol", size: 128, mode: os.FileMode(438), modTime: time.Unix(1625163935, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _stringNolol = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\x91\x31\x8f\xd4\x30\x10\x85\xeb\xcb\xaf\x78\x97\xe6\x58\x2d\x21\xa2\x3e\x16\x89\x92\x16\xd1\x23\xe3\x4c\x92\x91\x92\xb1\x35\x33\x39\xe5\xe7\x23\x27\x61\x51\xd0\x52\x50\x3e\xfb\x79\xde\xf7\x3c\x6d\x8b\xef\x23\x1b\x7a\x9e\x08\x31\x89\x07\x16\xc3\xcf\x60\x1c\xd1\x51\xcf\xc2\xce\x49\x0c\x41\x3a\xcc\x21\x6a\x32\xf4\x49\x61\xae\x2c\x43\x33\x07\xe1\xbc\x4c\xa1\x78\xaa\xb6\xc5\xd7\x39\x27\x75\x2c\xc6\x32\xe0\x05\x2c\x71\x5a\x3a\x42\x6d\xde\xb5\xfb\x93\x1a\x2f\x55\x71\x7e\xa3\x39\xbd\x91\xc1\x47\xc2\x14\xcc\x11\xc7\xa0\x21\x3a\x29\x7a\x4d\x73\x09\xd8\x32\xf3\x14\x22\x19\xd8\xc1\xe2\x09\x69\xf1\x6a\xc3\x38\x08\x7e\xe4\x94\xdf\x99\xeb\xfb\x72\x73\xf9\xc4\x83\x24\xa5\xcf\x98\x58\xa8\x7a\xda\xd5\xcd\x5c\x5f\xcb\xf5\x6d\xd7\x8d\xb9\x36\x0d\x2a\x92\xee\x20\xf1\x45\xc5\xf0\x11\xdc\x6f\xb1\xf7\x5f\x58\xcf\x51\xbf\xcf\xf7\xbc\xf5\x02\x5a\xb3\x56\x4f\x45\x36\xeb\xe5\xb9\xe4\xdc\x87\x7e\xe9\xba\xa3\x1b\xc9\x30\x3a\xd2\x3e\xfa\x68\xd0\xb6\x9b\x62\x83\x91\x97\xc3\xba\x06\xcb\xe6\xcf\x9a\x22\x99\x9d\x93\x27\x92\x3f\x25\x8f\x72\x23\x95\xa2\x3b\xf2\xf3\xad\xae\xcb\x6b\xc1\xde\x6d\xab\x7b\xbd\xe2\x15\x43\xf2\x84\x62\x45\x21\xbb\xd3\xe5\x4c\x72\x00\x2a\xbd\x91\x1a\x9d\x09\x3f\xfc\x37\xe2\x31\xe6\x9f\xbb\x78\x8c\xfb\xd7\x86\xae\xe7\x15\x3d\xc0\xff\x15\x00\x00\xff\xff\xf2\xf5\x00\x01\xaf\x02\x00\x00")

func stringNololBytes() ([]byte, error) {
	return bindataRead(
		_stringNolol,
		"string.nolol",
	)
}

func stringNolol() (*asset, error) {
	bytes, err := stringNololBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "string.nolol", size: 687, mode: os.FileMode(438), modTime: time.Unix(1623961616, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"logic.nolol":             logicNolol,
	"math_advanced.nolol":     math_advancedNolol,
	"math_basic.nolol":        math_basicNolol,
	"math_professional.nolol": math_professionalNolol,
	"string.nolol":            stringNolol,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"logic.nolol":             &bintree{logicNolol, map[string]*bintree{}},
	"math_advanced.nolol":     &bintree{math_advancedNolol, map[string]*bintree{}},
	"math_basic.nolol":        &bintree{math_basicNolol, map[string]*bintree{}},
	"math_professional.nolol": &bintree{math_professionalNolol, map[string]*bintree{}},
	"string.nolol":            &bintree{stringNolol, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

func assetFS() *assetfs.AssetFS {
	assetInfo := func(path string) (os.FileInfo, error) {
		return os.Stat(path)
	}
	for k := range _bintree.Children {
		return &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: assetInfo, Prefix: k}
	}
	panic("unreachable")
}
