package libs

// #include <string.h>
// #include "sass/context.h"
//
// extern struct Sass_Import** HeaderBridge(int idx);
//
// Sass_Import_List SassHeaders(const char* cur_path, Sass_Importer_Entry cb, struct Sass_Compiler* comp)
// {
//   void* cookie = sass_importer_get_cookie(cb);
//   int idx = *((int *)cookie);
//   Sass_Import_List list = HeaderBridge(idx);
//   return list;
//
// }
//
import "C"
import "unsafe"

var globalHeaders SafeMap

func init() {
	globalHeaders.init()
}

// BindHeader attaches the header to a libsass context ensuring
// gc does not delete the pointers necessary to make this happen.
func BindHeader(opts SassOptions, entries []ImportEntry) int {

	idx := globalHeaders.Set(entries)
	// ptr := unsafe.Pointer(idx)
	czero := C.double(0)
	imper := C.sass_make_importer(
		C.Sass_Importer_Fn(C.SassHeaders),
		czero,
		unsafe.Pointer(idx),
	)
	impers := C.sass_make_importer_list(1)
	C.sass_importer_set_list_entry(impers, 0, imper)

	C.sass_option_set_c_headers(
		(*C.struct_Sass_Options)(unsafe.Pointer(opts)),
		impers)
	return *idx
}

func RemoveHeaders(idx int) error {
	globalHeaders.Del(idx)
	return nil
}
