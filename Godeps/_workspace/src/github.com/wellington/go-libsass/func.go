package context

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

import (
	"reflect"
	"unsafe"
)

// SassCallback defines the callback libsass eventually executes in
// sprite_sass
type SassCallback func(ctx *Context, csv UnionSassValue) UnionSassValue

// Cookie is used for passing context information to libsass.  Cookie is
// passed to custom handlers when libsass executes them through the go
// bridge.
type Cookie struct {
	Sign string
	Fn   SassCallback
	Ctx  *Context
}

type handler struct {
	sign     string
	callback func(ctx *Context, csv UnionSassValue) UnionSassValue
}

// handlers is the list of registered sass handlers
var handlers []handler

// RegisterHandler sets the passed signature and callback to the
// handlers array.
func RegisterHandler(sign string,
	callback func(ctx *Context, csv UnionSassValue) UnionSassValue) {
	handlers = append(handlers, handler{sign, callback})
}

func (ctx *Context) SetFunc(opts *C.struct_Sass_Options) {
	cookies := make([]Cookie, len(handlers)+len(ctx.Cookies))
	// Append registered handlers to cookie array
	for i, h := range handlers {
		cookies[i] = Cookie{
			h.sign, h.callback, ctx,
		}
	}
	for i, h := range ctx.Cookies {
		cookies[i+len(handlers)] = Cookie{
			h.Sign, h.Fn, ctx,
		}
	}
	ctx.Cookies = cookies
	size := C.size_t(len(ctx.Cookies))
	fns := C.sass_make_function_list(size)

	// Send cookies to libsass
	// Create a slice that's backed by a C array
	length := len(ctx.Cookies) + 1
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(fns)),
		Len:  length, Cap: length,
	}

	gofns := *(*[]C.Sass_Function_Entry)(unsafe.Pointer(&hdr))
	for i, cookie := range ctx.Cookies {
		sign := C.CString(cookie.Sign)

		fn := C.sass_make_function(
			// sass signature
			sign,
			// C bridge
			C.Sass_Function_Fn(C.CallSassFunction),
			// Only pass reference to global array, so
			// GC won't clean it up.
			unsafe.Pointer(&ctx.Cookies[i]))

		gofns[i] = fn
	}
	C.sass_option_set_c_functions(opts, (C.Sass_Function_List)(unsafe.Pointer(&gofns[0])))

}
