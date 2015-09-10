package libs

// #include <stdio.h>
// #include <string.h>
// #include "sass/context.h"
//
// extern struct Sass_Import** HeaderBridge(void* cookie);
//
// Sass_Import_List SassHeaders(const char* cur_path, Sass_Importer_Entry cb, struct Sass_Compiler* comp)
// {
//   void* cookie = sass_importer_get_cookie(cb);
//   Sass_Import_List list = HeaderBridge(cookie);
//   return list;
//
// }
//
import "C"
import (
	"runtime"
	"unsafe"
)

// This binds the header to the libsass header lookup
func BindHeader(opts SassOptions, entries []ImportEntry) {
	ptr := unsafe.Pointer(&entries)
	// FIXME: this should be cleaned up manually later
	runtime.SetFinalizer(&entries, nil)
	imper := C.sass_make_importer(
		C.Sass_Importer_Fn(C.SassHeaders),
		C.double(0),
		ptr,
	)
	impers := C.sass_make_importer_list(1)
	C.sass_importer_set_list_entry(impers, 0, imper)

	C.sass_option_set_c_headers(
		(*C.struct_Sass_Options)(unsafe.Pointer(opts)),
		impers)
}
