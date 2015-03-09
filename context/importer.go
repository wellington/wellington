package context

// #include <stdlib.h>
// #include <string.h>
// #include "sass_context.h"
//
// extern struct Sass_Import** ImporterBridge(const char* url, const char* prev, void* cookie);
// struct Sass_Import** SassImporter(const char* url, const char* prev, void* cookie)
// {
//   struct Sass_Import** golist = ImporterBridge(url, prev, cookie);
//   const char* src = sass_import_get_source(golist[0]);
//   // printf("There should be code in this: %s\n", src);
//   return golist;
// }
//
// #ifndef UINTMAX_MAX
// #  ifdef __UINTMAX_MAX__
// #    define UINTMAX_MAX __UINTMAX_MAX__
// #  else
// #    error
// #  endif
// #endif
//
// size_t max_size = UINTMAX_MAX;
import "C"
import (
	"errors"
	"io"
	"io/ioutil"
	"sync"
	"time"
	"unsafe"
)

// SassImport wraps Sass_Import libsass struct
type SassImport C.struct_Sass_Import

// MaxSizeT is safe way of specifying size_t = -1
var MaxSizeT = C.max_size

var (
	ErrImportNotFound = errors.New("Import unreachable or not found")
)

// Import contains Rel and Abs path and a string of the contents
// representing an import.
type Import struct {
	Body  io.ReadCloser
	bytes []byte
	mod   time.Time
}

// ModTime returns modification time
func (i Import) ModTime() time.Time {
	return i.mod
}

// Imports is a map with key of "path/to/file"
type Imports struct {
	sync.RWMutex
	m map[string]Import
}

// Add registers an import in the context.Imports
func (p *Imports) Add(path string, rc io.ReadCloser) error {
	p.Lock()
	defer p.Unlock()

	bs, err := ioutil.ReadAll(rc)
	if err != nil {
		return err
	}
	im := Import{
		Body:  rc,
		bytes: bs,
		mod:   time.Now(),
	}
	p.m[path] = im
	return nil
}

// Del removes the import from the context.Imports
func (p *Imports) Del(path string) {
	p.Lock()
	defer p.Unlock()

	delete(p.m, path)
}

// Get retrieves import bytes by path
func (p *Imports) Get(path string) ([]byte, error) {
	p.RLock()
	defer p.RUnlock()

	imp, ok := p.m[path]
	if !ok {
		return nil, ErrImportNotFound
	}
	return imp.bytes, nil
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

// SetImporter enables custom importer in libsass
func (ctx *Context) SetImporter(opts *C.struct_Sass_Options) {
	if ctx.Imports.Len() == 0 {
		return
	}
	p := C.Sass_C_Import_Fn(C.SassImporter)
	impCallback := C.sass_make_importer(p,
		unsafe.Pointer(ctx))
	C.sass_option_set_importer(opts, impCallback)
}
