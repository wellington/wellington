package context

// #include "sass_context.h"
// #include "sass_functions.h"
import "C"
import (
	"fmt"
	"testing"
)

// SassImport ...
type SassImport C.struct_Sass_Import

// ImportCallback ...
type ImportCallback C.Sass_C_Import_Callback

func testSassImport(t *testing.T) {
	/*struct Sass_Import** list = sass_make_import_list(2);
	const char* local = "local { color: green; }";
	const char* remote = "remote { color: red; }";
	list[0] = sass_make_import_entry("/tmp/styles.scss", strdup(local), 0);
	list[1] = sass_make_import_entry("http://www.example.com", strdup(remote), 0);
	return list;*/
	var entries []*C.struct_Sass_Import
	entry := C.sass_make_import_entry(C.CString("a"), C.CString("a { color: red; }"), C.CString(""))
	entries = append(entries, entry)
	path := C.sass_import_get_path(entries[0])
	fmt.Println(C.GoString(path))

	//fmt.Println(entry.)
}
