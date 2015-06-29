package libs

// #include <stdlib.h>
// #include "sass_context.h"
//
// extern union Sass_Value* GoBridge( union Sass_Value* s_args, void* cookie);
//
// union Sass_Value* CallSassFunction( union Sass_Value* s_args, Sass_Function_Entry cb, struct Sass_Options* opts ) {
//     void* cookie = sass_function_get_cookie(cb);
//     return GoBridge(s_args, cookie);
// }
//
import "C"
import "unsafe"

type SassFunc C.Sass_Function_Entry

func SassMakeFunction(signature string, ptr unsafe.Pointer) SassFunc {
	csign := C.CString(signature)
	ck := *(*Cookie)(ptr)
	_ = ck
	fn := C.sass_make_function(csign,
		C.Sass_Function_Fn(C.CallSassFunction),
		ptr)
	return (SassFunc)(fn)
}

func BindFuncs(opts SassOptions, funcs []SassFunc) {
	cfuncs := C.sass_make_function_list(C.size_t(len(funcs)))
	for i, cfn := range funcs {
		C.sass_function_set_list_entry(cfuncs, C.size_t(i), cfn)
	}
	C.sass_option_set_c_functions(opts, cfuncs)
}
