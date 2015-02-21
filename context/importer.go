package context

// #include <stdlib.h>
// #include <string.h>
// #include <stdio.h>
// #include "sass_functions.h"
// #include "sass_context.h"
//
// extern struct Sass_Import** ImporterBridge(const char* url, const char* prev, void* cookie);
// struct Sass_Import** SassImporter(const char* url, const char* prev, void* cookie)
// {
//   //printf("sass_importer url: %s prev: %s\n", url, prev);
//   /*struct Sass_Import** list = sass_make_import_list(2);
//   const char* local = "local { color: green; }";
//   const char* remote = "remote { color: red; }";
//   list[0] = sass_make_import_entry("/tmp/styles.scss", strdup(local), 0);
//   list[1] = sass_make_import_entry("http://www.example.com", strdup(remote), 0);*/
//
//   return ImporterBridge(url, prev, cookie);
// }
//
import "C"
import (
	"bytes"
	"fmt"
	"testing"
	"unsafe"
)

// SassImport ...
type SassImport C.struct_Sass_Import

// ImportCallback ...
type ImportCallback C.Sass_C_Import_Callback

type Import struct {
	Rel      string
	Abs      string
	Contents string
}

func (ctx *Context) AddImport(name string, contents string) {
	ctx.Imports = append(ctx.Imports, Import{
		Rel:      name,
		Contents: contents,
	})
}

func (ctx *Context) FindImport(name string) (Import, bool) {
	for i := range ctx.Imports {
		if ctx.Imports[i].Rel == name {
			return ctx.Imports[i], true
		}
	}
	return Import{}, false
}

func (ctx *Context) SetImporter(opts *C.struct_Sass_Options) {
	p := C.Sass_C_Import_Fn(C.SassImporter)
	impCallback := C.sass_make_importer(p,
		unsafe.Pointer(ctx))
	C.sass_option_set_importer(opts, impCallback)
}

func testSassImport(t *testing.T) {

	in := bytes.NewBufferString(`@import "a";`)

	var out bytes.Buffer
	ctx := Context{}
	ctx.AddImport("a", "a { color: blue; }")
	err := ctx.Compile(in, &out)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("out", out.String())

	/*var entries []*SassImport
	entry := C.sass_make_import_entry(
		C.CString("a"),
		C.CString("a { color: red; }"),
		C.CString(""))
	entries = append(entries, (*SassImport)(entry))
	path := C.sass_import_get_path((*C.struct_Sass_Import)(unsafe.Pointer(&entries[0])))
	fmt.Println(C.GoString(path))
	path = C.sass_import_get_source((*C.struct_Sass_Import)(entries[0]))
	fmt.Println(C.GoString(path))*/

}
