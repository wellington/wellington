package context

// #include <stdlib.h>
// #include <stdio.h>
// #include "sass_functions.h"
// #include "sass_context.h"
//
//
import "C"
import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"
)

// SassImport ...
type SassImport C.struct_Sass_Import

// ImportCallback ...
type ImportCallback C.Sass_C_Import_Callback

// AddImport ...
func (ctx *Context) AddImport(name string, contents string) {
	path := C.CString(name)
	cnts := C.CString(contents)
	empty := C.CString("")
	//defer C.free(unsafe.Pointer(path))
	//defer C.free(unsafe.Pointer(cnts))
	//defer C.free(unsafe.Pointer(empty))
	//defer C.free(unsafe.Pointer(abss))
	abs, _ := filepath.Abs(name)
	abss := C.CString(abs)
	entry := C.sass_make_import(path, abss, cnts, empty)
	//ctx.Imports = append(ctx.Imports, (*SassImport)(entry))
	ctx.Imports = append(ctx.Imports, entry)
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

	src := C.sass_import_get_source((*C.struct_Sass_Import)(ctx.Imports[0]))
	fmt.Printf("%s\n", C.GoString(src))

	fmt.Println(out.String())

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
