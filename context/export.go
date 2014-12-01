package context

// The use of //export in context.go prevents being able to define any C code in the preamble of that file.  Export defines additional C code necessary for the context<->sass_context bridge.
// See: http://golang.org/cmd/cgo/#hdr-C_references_to_Go

/*
#cgo LDFLAGS: -lsass -lstdc++ -lm
#cgo CFLAGS:

#include "sass_context.h"

extern union Sass_Value* customHandler( union Sass_Value* s_args, void* cookie);

union Sass_Value* CallSassFunction( union Sass_Value* s_args, void* cookie ) {
    // printf("callback yo\n");
    // union Sass_Value* sass_value = NULL;
    // int a;
    // return sass_make_boolean(false);
    return customHandler(s_args, cookie);
}

*/
import "C"
