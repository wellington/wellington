package context

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
