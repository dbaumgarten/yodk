package stdlib

import "strings"

// Prefix is how all files of the standard library start
const Prefix = "std/"

// Is returns true if the given package-name belongs to the standard-library
func Is(name string) bool {
	return strings.HasPrefix(name, Prefix)
}

// Get returns the wanted file from the standard-library (or an error)
func Get(name string) (string, error) {
	name = trim(name)
	by, err := Asset(name)
	if err != nil {
		return "", err
	}
	return string(by), err
}

// Trim removes the prefix from the path
func trim(name string) string {
	return strings.TrimPrefix(name, Prefix)
}
