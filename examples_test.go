package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	yodktesting "github.com/dbaumgarten/yodk/pkg/testing"
)

func TestExamples(t *testing.T) {
	runTestfiles(filepath.Join("examples", "yolol"), t)
	runTestfiles(filepath.Join("examples", "nolol"), t)
}

func runTestfiles(path string, t *testing.T) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}
	testfiles := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), "_test.yaml") {
			testfiles = append(testfiles, filepath.Join(path, file.Name()))
		}
	}

	for _, testfile := range testfiles {
		t.Run(testfile, func(t *testing.T) {
			file, err := ioutil.ReadFile(testfile)
			if err != nil {
				t.Fatal(err)
			}
			absolutePath, _ := filepath.Abs(testfile)
			test, err := yodktesting.Parse([]byte(file), absolutePath)
			if err != nil {
				t.Fatal(err)
			}
			fails := test.Run(nil)
			if len(fails) != 0 {
				for _, fail := range fails {
					t.Log(fail)
				}
				t.FailNow()
			}
		})
	}
}
