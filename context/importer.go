package context

// #include <stdlib.h>
// #include <string.h>
// #include <stdio.h>
// #include "sass_functions.h"
// #include "sass_context.h"
//
// extern struct Sass_Import** ImporterBridge(const char* url, const char* prev, void* cookie);
// struct Sass_Import** SassImporter(const char* url, const char* prev, void* cookie)
// {
//   struct Sass_Import** golist = ImporterBridge(url, prev, cookie);
//   const char* src = sass_import_get_source(golist[0]);
//   // printf("There should be code in this: %s\n", src);
//   return golist;
// }
//
import "C"
import "unsafe"

// SassImport ...
type SassImport C.struct_Sass_Import

// ImportCallback ...
type ImportCallback C.Sass_C_Import_Callback

type Import struct {
	Rel      string
	Abs      string
	Contents string
}

func (ctx *Context) AddImport(name string, contents string) {
	ctx.Imports = append(ctx.Imports, Import{
		Rel:      name,
		Contents: contents,
	})
}

func (ctx *Context) FindImport(name string) (Import, bool) {
	for i := range ctx.Imports {
		if ctx.Imports[i].Rel == name {
			return ctx.Imports[i], true
		}
	}
	return Import{}, false
}

func (ctx *Context) SetImporter(opts *C.struct_Sass_Options) {
	if len(ctx.Imports) == 0 {
		return
	}
	p := C.Sass_C_Import_Fn(C.SassImporter)
	impCallback := C.sass_make_importer(p,
		unsafe.Pointer(ctx))
	C.sass_option_set_importer(opts, impCallback)
}
