package libs

// #include <stdint.h>
// #include <stdlib.h>
// #include <string.h>
// #include "sass_context_bind.h"
//
// extern struct Sass_Import** ImporterBridge(const char* url, const char* prev, void* cookie);
//
// Sass_Import_List SassImporterHandler(const char* cur_path, Sass_Importer_Entry cb, struct Sass_Compiler* comp)
// {
//   void* cookie = sass_importer_get_cookie(cb);
//   struct Sass_Import* previous = sass_compiler_get_last_import(comp);
//   const char* prev_path = sass_import_get_path(previous);
//   Sass_Import_List list = ImporterBridge(cur_path, prev_path, cookie);
//   return list;
// }
//
//
// #ifndef UINTMAX_MAX
// #  ifdef __UINTMAX_MAX__
// #    define UINTMAX_MAX __UINTMAX_MAX__
// #  endif
// #endif
//
// //size_t max_size = UINTMAX_MAX;
import "C"
import (
	"runtime"
	"unsafe"
)

func BindImporter(opts SassOptions, entries []ImportEntry) {
	ptr := unsafe.Pointer(&entries)
	runtime.SetFinalizer(&entries, nil)

	imper := C.sass_make_importer(
		C.Sass_Importer_Fn(C.SassImporterHandler),
		C.double(0),
		ptr,
	)
	impers := C.sass_make_importer_list(1)
	C.sass_importer_set_list_entry(impers, 0, imper)

	C.sass_option_set_c_importers(
		(*C.struct_Sass_Options)(unsafe.Pointer(opts)),
		impers,
	)
}
