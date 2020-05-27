package debug

import (
	"os"
)

// StdioReadWriteCloser is a ReadWriteCloser for reading from stdin and writing to stdout
// If Log is true, all incoming and outgoing data is logged
type StdioReadWriteCloser struct {
}

func (s StdioReadWriteCloser) Read(p []byte) (n int, err error) {
	return os.Stdin.Read(p)
}

func (s StdioReadWriteCloser) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

// Close closes stdin and stdout
func (s StdioReadWriteCloser) Close() error {
	err := os.Stdin.Close()
	if err != nil {
		return err
	}
	return os.Stdout.Close()
}
