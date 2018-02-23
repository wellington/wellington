package libs

// #include <stdlib.h>
// #include "sass/context.h"
import "C"

import (
	"reflect"
	"unsafe"
)

func GetImportList(ctx SassContext) []string {
	cctx := (*C.struct_Sass_Context)(ctx)
	len := int(C.sass_context_get_included_files_size(cctx))
	imps := C.sass_context_get_included_files(cctx)
	list := make([]string, len, len)
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(imps)),
		Len:  len, Cap: len,
	}
	goimps := *(*[]*C.char)(unsafe.Pointer(&hdr))
	for i := range goimps {
		list[i] = C.GoString(goimps[i])
	}
	return list
}

func Version() string {
	s := C.libsass_version()
	return C.GoString(s)
}
