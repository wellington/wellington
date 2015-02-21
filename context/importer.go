package context

// #include "sass_context.h"
// #include "sass_functions.h"
import "C"
import (
	"fmt"
	"testing"
	"unsafe"
)

// SassImport ...
type SassImport C.struct_Sass_Import

// ImportCallback ...
type ImportCallback C.Sass_C_Import_Callback

func testSassImport(t *testing.T) {

	var entries []*SassImport
	entry := C.sass_make_import_entry(
		C.CString("a"),
		C.CString("a { color: red; }"),
		C.CString(""))
	entries = append(entries, (*SassImport)(entry))
	path := C.sass_import_get_path((*C.struct_Sass_Import)(unsafe.Pointer(&entries[0])))
	fmt.Println(C.GoString(path))
	path = C.sass_import_get_source((*C.struct_Sass_Import)(entries[0]))
	fmt.Println(C.GoString(path))

}
