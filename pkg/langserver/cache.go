package langserver

import (
	"fmt"
	"sync"

	"github.com/dbaumgarten/yodk/pkg/lsp"
	"github.com/dbaumgarten/yodk/pkg/nolol"
)

var NotFoundError = fmt.Errorf("File not found in cache")

type Cache struct {
	Files       map[lsp.DocumentURI]string
	Diagnostics map[lsp.DocumentURI]DiagnosticResults
	Lock        *sync.Mutex
}

type DiagnosticResults struct {
	Variables      []string
	AnalysisReport *nolol.AnalysisReport
}

func NewCache() *Cache {
	return &Cache{
		Files:       make(map[lsp.DocumentURI]string),
		Diagnostics: make(map[lsp.DocumentURI]DiagnosticResults),
		Lock:        &sync.Mutex{},
	}
}

func (c *Cache) Get(uri lsp.DocumentURI) (string, error) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	f, found := c.Files[uri]
	if !found {
		return "", NotFoundError
	}
	return f, nil
}

func (c *Cache) Set(uri lsp.DocumentURI, content string) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	c.Files[uri] = content
}

func (c *Cache) GetDiagnostics(uri lsp.DocumentURI) (*DiagnosticResults, error) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	f, found := c.Diagnostics[uri]
	if !found {
		return nil, NotFoundError
	}
	return &f, nil
}

func (c *Cache) SetDiagnostics(uri lsp.DocumentURI, content DiagnosticResults) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	c.Diagnostics[uri] = content
}
