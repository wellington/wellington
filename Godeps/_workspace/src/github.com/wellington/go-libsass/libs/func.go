package libs

// #include <stdlib.h>
// #include "sass_context_bind.h"
//
// extern union Sass_Value* GoBridge( union Sass_Value* s_args, void* cookie);
//
// union Sass_Value* CallSassFunction( union Sass_Value* s_args, Sass_Function_Entry cb, struct Sass_Options* opts ) {
//     void* cookie = sass_function_get_cookie(cb);
//     union Sass_Value* ret;
//     ret = GoBridge(s_args, cookie);
//     return ret;
// }
//
import "C"
import "unsafe"

type SassFunc C.Sass_Function_Entry

func SassMakeFunction(signature string, ptr unsafe.Pointer) SassFunc {
	csign := C.CString(signature)
	fn := C.sass_make_function(
		csign,
		C.Sass_Function_Fn(C.CallSassFunction),
		ptr)
	return (SassFunc)(fn)
}

func BindFuncs(opts SassOptions, funcs []SassFunc) {
	sz := C.size_t(len(funcs))
	cfuncs := C.sass_make_function_list(sz)
	for i, cfn := range funcs {
		C.sass_function_set_list_entry(cfuncs, C.size_t(i), cfn)
	}
	C.sass_option_set_c_functions(opts, cfuncs)
}
