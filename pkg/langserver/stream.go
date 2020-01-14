package langserver

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type StdioStream struct {
	in    io.Reader
	out   io.Writer
	bufin *bufio.Reader
	lock  *sync.Mutex
	Log   bool
}

func NewStdioStream() *StdioStream {
	return &StdioStream{
		in:    os.Stdin,
		out:   os.Stdout,
		bufin: bufio.NewReader(os.Stdin),
		lock:  &sync.Mutex{},
	}
}

func (s *StdioStream) Write(ctx context.Context, data []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.Log {
		log.Printf("Sent: %s", string(data))
	}
	if _, err := fmt.Fprintf(s.out, "Content-Length: %d\r\n\r\n", len(data)); err != nil {
		return err
	}
	if _, err := s.out.Write(data); err != nil {
		return err
	}
	return nil
}

// ReadObject implements ObjectCodec.
func (s *StdioStream) Read(ctx context.Context) ([]byte, error) {
	var contentLength uint64
	for {
		line, err := s.bufin.ReadString('\r')
		if err != nil {
			return nil, err
		}
		b, err := s.bufin.ReadByte()
		if err != nil {
			return nil, err
		}
		if b != '\n' {
			return nil, fmt.Errorf(`jsonrpc2: line endings must be \r\n`)
		}
		if line == "\r" {
			break
		}
		if strings.HasPrefix(line, "Content-Length: ") {
			line = strings.TrimPrefix(line, "Content-Length: ")
			line = strings.TrimSpace(line)
			var err error
			contentLength, err = strconv.ParseUint(line, 10, 32)
			if err != nil {
				return nil, err
			}
		}
	}
	if contentLength == 0 {
		return nil, fmt.Errorf("jsonrpc2: no Content-Length header found")
	}

	result, err := ioutil.ReadAll(io.LimitReader(s.bufin, int64(contentLength)))
	if err != nil {
		return nil, err
	}
	if s.Log {
		log.Printf("Received: %s", string(result))
	}
	return result, err
}
