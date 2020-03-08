package langserver

import (
	"fmt"
	"sync"

	"github.com/dbaumgarten/yodk/pkg/lsp"
)

var NotFoundError = fmt.Errorf("File not found in cache")

type Cache struct {
	Files map[lsp.DocumentURI]string
	Lock  *sync.Mutex
}

func NewCache() *Cache {
	return &Cache{
		Files: make(map[lsp.DocumentURI]string),
		Lock:  &sync.Mutex{},
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
