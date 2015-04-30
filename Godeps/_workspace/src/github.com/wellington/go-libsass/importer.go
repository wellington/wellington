package context

// #include <stdint.h>
// #include <stdlib.h>
// #include <string.h>
// #include "sass_context.h"
//
// extern struct Sass_Import** ImporterBridge(const char* url, const char* prev, void* cookie);
//
// Sass_Import_List SassImporter(const char* cur_path, Sass_Importer_Entry cb, struct Sass_Compiler* comp)
// {
//   void* cookie = sass_importer_get_cookie(cb);
//   struct Sass_Import* previous = sass_compiler_get_last_import(comp);
//   const char* prev_path = sass_import_get_path(previous);
//   Sass_Import_List list = ImporterBridge(cur_path, prev_path, cookie);
//   return list;
// }
//
// #ifndef UINTMAX_MAX
// #  ifdef __UINTMAX_MAX__
// #    define UINTMAX_MAX __UINTMAX_MAX__
// #  endif
// #endif
//
// size_t max_size = UINTMAX_MAX;
import "C"
import (
	"errors"
	"io"
	"reflect"
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

// Init sets up a new Imports map
func (p *Imports) Init() {
	p.m = make(map[string]Import)
}

// Add registers an import in the context.Imports
func (p *Imports) Add(prev string, cur string, bs []byte) error {
	p.Lock()
	defer p.Unlock()

	im := Import{
		bytes: bs,
		mod:   time.Now(),
	}
	// TODO: align these with libsass name "stdin"
	if len(prev) == 0 || prev == "string" {
		prev = "stdin"
	}
	p.m[prev+":"+cur] = im
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
	imp, ok := p.m[prev+":"+path]
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

	imps := C.sass_make_importer_list(1)
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(imps)),
		Len:  1, Cap: 1,
	}
	goimps := *(*[]C.Sass_Importer_Entry)(unsafe.Pointer(&hdr))
	p := C.SassImporter
	imp := C.sass_make_importer(
		C.Sass_Importer_Fn(p),
		C.double(0),
		unsafe.Pointer(ctx))
	goimps[0] = imp
	C.sass_option_set_c_importers(opts, imps)
}
