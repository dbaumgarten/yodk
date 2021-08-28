package stdlib

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestGeneratedCode(t *testing.T) {

	files := AssetNames()
	filesOnDisk, err := ioutil.ReadDir("src")
	if err != nil {
		t.Error(err)
	}

	if len(files) != len(filesOnDisk) {
		t.Fatal("Amount of files in bindata does not match files on disk. You need to re-run go-bindata-assetfs")
	}

	for _, name := range files {
		bindataContent, _ := Asset(name)
		fileContent, err := ioutil.ReadFile(filepath.Join("src", name))
		if err != nil {
			t.Error(err)
		}
		if string(bindataContent) != string(fileContent) {
			t.Fatalf("File stdlib/src/%s has changed. You need to re-run go-bindata-assetfs", name)
		}
	}

}
