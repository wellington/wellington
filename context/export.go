package context

// The use of //export in context.go prevents being able to define any C code in the preamble of that file.  Export defines additional C code necessary for the context<->sass_context bridge.
// See: http://golang.org/cmd/cgo/#hdr-C_references_to_Go

/*
#cgo LDFLAGS: -lsass -lstdc++ -lm
#cgo CFLAGS:

#include "sass_context.h"

extern union Sass_Value* customHandler( union Sass_Value* s_args, void* cookie);

union Sass_Value* CallSassFunction( union Sass_Value* s_args, void* cookie ) {
    return customHandler(s_args, cookie);
}
*/
import "C"
import "log"

type Cookie struct {
	Lane int
	sign string
	fn   SassCallback
	ctx  *Context
}

type UnionSassValue *C.union_Sass_Value

// CookieCb defines the callback libsass eventually executes in sprite_sass
type SassCallback func(csv UnionSassValue) UnionSassValue

// NewCookie creates C.Cookie from Go strings.  It is not safe and will leak
// memory, so structs created need to be cleaned up manually.
func NewCookie(lane int, sign string) Cookie {
	var c Cookie
	c.Lane = lane
	c.sign = sign
	return c
}

func SampleCB(usv UnionSassValue) UnionSassValue {
	var sv []SassValue
	Unmarshal(usv, &sv)
	log.Printf("Arguments: % #v\n", sv)
	return C.sass_make_boolean(false)
}
