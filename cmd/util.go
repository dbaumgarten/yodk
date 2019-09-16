package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
)

var inputFile string

func exitOnError(err error, operation string) {
	if err != nil {
		fmt.Printf("Error when %s: %s\n", operation, err.Error())
		os.Exit(1)
	}
}

func loadInputFile(file string) string {
	f, err := ioutil.ReadFile(file)
	exitOnError(err, "Loading input file")
	return string(f)
}
