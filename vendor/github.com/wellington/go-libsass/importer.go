package libsass

import (
	"errors"
	"io"
	"sync"
	"time"

	"github.com/wellington/go-libsass/libs"
)

var (
	ErrImportNotFound = errors.New("Import unreachable or not found")
)

// Import contains Rel and Abs path and a string of the contents
// representing an import.
type Import struct {
	Body  io.ReadCloser
	bytes []byte
	mod   time.Time
	Prev  string
	Path  string
}

// ModTime returns modification time
func (i Import) ModTime() time.Time {
	return i.mod
}

// Imports is a map with key of "path/to/file"
type Imports struct {
	wg      sync.WaitGroup
	closing chan struct{}
	sync.RWMutex
	m   map[string]Import
	idx int
}

func NewImports() *Imports {
	return &Imports{
		closing: make(chan struct{}),
	}
}

func (i *Imports) Close() {
	close(i.closing)
	i.wg.Wait()
}

// Init sets up a new Imports map
func (p *Imports) Init() {
	p.m = make(map[string]Import)
}

// Add registers an import in the context.Imports
func (p *Imports) Add(prev string, path string, bs []byte) error {
	p.Lock()
	defer p.Unlock()

	// TODO: align these with libsass name "stdin"
	if len(prev) == 0 || prev == "string" {
		prev = "stdin"
	}
	im := Import{
		bytes: bs,
		mod:   time.Now(),
		Prev:  prev,
		Path:  path,
	}

	p.m[prev+":"+path] = im
	return nil
}

// Del removes the import from the context.Imports
func (p *Imports) Del(path string) {
	p.Lock()
	defer p.Unlock()

	delete(p.m, path)
}

// Get retrieves import bytes by path
func (p *Imports) Get(prev, path string) ([]byte, error) {
	p.RLock()
	defer p.RUnlock()
	for _, imp := range p.m {
		if imp.Prev == prev && imp.Path == path {
			return imp.bytes, nil
		}
	}
	return nil, ErrImportNotFound
}

// Update attempts to create a fresh Body from the given path
// Files last modified stamps are compared against import timestamp
func (p *Imports) Update(name string) {
	p.Lock()
	defer p.Unlock()

}

// Len counts the number of entries in context.Imports
func (p *Imports) Len() int {
	return len(p.m)
}

// Bind accepts a SassOptions and adds the registered
// importers in the context.
func (p *Imports) Bind(opts libs.SassOptions) {
	entries := make([]libs.ImportEntry, p.Len())
	i := 0

	p.RLock()
	for _, ent := range p.m {
		bs := ent.bytes
		entries[i] = libs.ImportEntry{
			Parent: ent.Prev,
			Path:   ent.Path,
			Source: string(bs),
		}
		i++
	}
	p.RUnlock()

	// set entries somewhere so GC doesn't collect it
	p.idx = libs.BindImporter(opts, entries)
}
