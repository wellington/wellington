package libs

// extern struct Sass_Import** HeaderBridge(int idx);
//
//
// #//for C.free
// #include "stdlib.h"
//
// #include "sass/context.h"
//
import "C"

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

type SassImporter *C.struct_Sass_Importer
type SassImporterList C.Sass_Importer_List

// SassMakeImporterList maps to C.sass_make_importer_list
func SassMakeImporterList(gol int) SassImporterList {
	l := C.size_t(gol)
	cimp := C.sass_make_importer_list(l)
	return (SassImporterList)(cimp)
}

type ImportEntry struct {
	Parent string
	Path   string
	Source string
	SrcMap string
}

//export HeaderBridge
func HeaderBridge(cint C.int) C.Sass_Import_List {
	idx := int(cint)
	entries, ok := globalHeaders.Get(idx).([]ImportEntry)
	if !ok {
		fmt.Printf("failed to resolve header slice: %d\n", idx)
		return C.sass_make_import_list(C.size_t(1))
	}

	cents := C.sass_make_import_list(C.size_t(len(entries)))

	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cents)),
		Len:  len(entries), Cap: len(entries),
	}
	goents := *(*[]C.Sass_Import_Entry)(unsafe.Pointer(&hdr))

	for i, ent := range entries {
		// Each entry needs a unique name
		cent := C.sass_make_import_entry(
			C.CString(ent.Path),
			C.CString(ent.Source),
			C.CString(ent.SrcMap))
		// There is a function for modifying an import list, but a proper
		// slice might be more useful.
		// C.sass_import_set_list_entry(cents, C.size_t(i), cent)
		goents[i] = cent
	}

	return cents
}

func GetEntry(es []ImportEntry, parent string, path string) (string, error) {
	for _, e := range es {
		if parent == e.Parent && path == e.Path {
			return e.Source, nil
		}
	}
	return "", errors.New("entry not found")
}

// ImporterBridge is called by C to pass Importer arguments into Go land. A
// Sass_Import is returned for libsass to resolve.
//
//export ImporterBridge
func ImporterBridge(url *C.char, prev *C.char, cidx C.int) C.Sass_Import_List {
	// Retrieve the index
	idx := int(cidx)
	entries, ok := globalImports.Get(idx).([]ImportEntry)
	if !ok {
		fmt.Printf("failed to resolve import slice: %d\n", idx)
		entries = []ImportEntry{}
	}

	parent := C.GoString(prev)
	rel := C.GoString(url)
	list := C.sass_make_import_list(1)
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(list)),
		Len:  1, Cap: 1,
	}

	golist := *(*[]C.Sass_Import_Entry)(unsafe.Pointer(&hdr))
	if body, err := GetEntry(entries, parent, rel); err == nil {
		ent := C.sass_make_import_entry(url, C.CString(body), nil)
		cent := (C.Sass_Import_Entry)(ent)
		golist[0] = cent
	} else if strings.HasPrefix(rel, "compass") {
		ent := C.sass_make_import_entry(url, C.CString(""), nil)
		cent := (C.Sass_Import_Entry)(ent)
		golist[0] = cent
	} else {
		ent := C.sass_make_import_entry(url, nil, nil)
		cent := (C.Sass_Import_Entry)(ent)
		golist[0] = cent
	}

	return list
}

type SassImportList C.Sass_Import_List

type SassFileContext *C.struct_Sass_File_Context

// SassMakeFileContext maps to C.sass_make_file_context
func SassMakeFileContext(gos string) SassFileContext {
	s := C.CString(gos)
	fctx := C.sass_make_file_context(s)
	return (SassFileContext)(fctx)
}

// SassDeleteFileContext frees memory used for a file context
func SassDeleteFileContext(fc SassFileContext) {
	C.sass_delete_file_context(fc)
}

type SassDataContext *C.struct_Sass_Data_Context

// SassMakeDataContext creates a data context from a source string
func SassMakeDataContext(gos string) SassDataContext {
	s := C.CString(gos)
	dctx := C.sass_make_data_context(s)
	return (SassDataContext)(dctx)
}

// SassDeleteDataContext frees the memory used for a data context
func SassDeleteDataContext(dc SassDataContext) {
	C.sass_delete_data_context(dc)
}

type SassContext *C.struct_Sass_Context

// SassDataContextGetContext returns a context from a data context.
// These are useful for calling generic methods like compiling.
func SassDataContextGetContext(godc SassDataContext) SassContext {
	opts := C.sass_data_context_get_context(godc)
	return (SassContext)(opts)
}

// SassFileContextGetContext returns a context from a file context.
// These are useful for calling generic methods like compiling.
func SassFileContextGetContext(gofc SassFileContext) SassContext {
	opts := C.sass_file_context_get_context(gofc)
	return (SassContext)(opts)
}

// SassOptions is a wrapper to C.struct_Sass_Options
type SassOptions *C.struct_Sass_Options

// SassMakeOptions creates a new SassOptions object
func SassMakeOptions() SassOptions {
	opts := C.sass_make_options()
	return (SassOptions)(opts)
}

// SassFileContextGetOptions returns the sass options in a file context
func SassFileContextGetOptions(gofc SassFileContext) SassOptions {
	fcopts := C.sass_file_context_get_options(gofc)
	return (SassOptions)(fcopts)
}

// SassFileContextGetOptions sets a sass options to a file context
func SassFileContextSetOptions(gofc SassFileContext, goopts SassOptions) {
	C.sass_file_context_set_options(gofc, goopts)
}

// SassDataContextGetOptions returns the Sass options in a data context
func SassDataContextGetOptions(godc SassDataContext) SassOptions {
	dcopts := C.sass_data_context_get_options(godc)
	return (SassOptions)(dcopts)
}

// SassDataContextGetOptions sets a Sass options to a data context
func SassDataContextSetOptions(godc SassDataContext, goopts SassOptions) {
	C.sass_data_context_set_options(godc, goopts)
}

type SassCompiler *C.struct_Sass_Compiler

// SassMakeFileCompiler returns a compiler from a file context
func SassMakeFileCompiler(gofc SassFileContext) SassCompiler {
	sc := C.sass_make_file_compiler(gofc)
	return (SassCompiler)(sc)
}

// SassMakeDataCompiler returns a compiler from a data context
func SassMakeDataCompiler(godc SassDataContext) SassCompiler {
	dc := C.sass_make_data_compiler(godc)
	return (SassCompiler)(dc)
}

// SassCompileFileContext compile from file context
func SassCompileFileContext(gofc SassFileContext) int {
	cstatus := C.sass_compile_file_context(gofc)
	return int(cstatus)
}

// SassCompilerParse prepares a compiler for execution
func SassCompilerParse(c SassCompiler) {
	C.sass_compiler_parse(c)
}

// SassCompilerExecute compiles a compiler
func SassCompilerExecute(c SassCompiler) {
	C.sass_compiler_execute(c)
}

// SassDeleteCompiler frees memory for the Sass compiler
func SassDeleteCompiler(c SassCompiler) {
	C.sass_delete_compiler(c)
}

// SassOptionSetCHeaders adds custom C headers to a SassOptions
func SassOptionSetCHeaders(gofc SassOptions, goimp SassImporterList) {
	C.sass_option_set_c_headers(gofc, C.Sass_Importer_List(goimp))
}

// SassContextGetOutputString retrieves the final compiled CSS after
// compiler parses and executes.
func SassContextGetOutputString(goctx SassContext) string {
	cstr := C.sass_context_get_output_string(goctx)
	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
}

// SassContextGetErrorJSON requests an error in JSON format from libsass
func SassContextGetErrorJSON(goctx SassContext) string {
	cstr := C.sass_context_get_error_json(goctx)
	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
}

// SassContextGetErrorStatus requests error status
func SassContextGetErrorStatus(goctx SassContext) int {
	return int(C.sass_context_get_error_status(goctx))
}

// SassOptionGetSourceMapFile retrieves the source map file
func SassOptionGetSourceMapFile(goopts SassOptions) string {
	p := C.sass_option_get_source_map_file(goopts)
	return C.GoString(p)
}

// SassContextGetSourceMapString retrieves the contents of a
// source map
func SassContextGetSourceMapString(goctx SassContext) string {
	s := C.sass_context_get_source_map_string(goctx)
	return C.GoString(s)
}

// SassOptionSetPrecision sets the precision of floating point math
// ie. 3.2222px. This is currently bugged and does not work.
func SassOptionSetPrecision(goopts SassOptions, i int) {
	C.sass_option_set_precision(goopts, C.int(i))
}

// SassOptionSetOutputStyle sets the output format of CSS see: http://godoc.org/github.com/wellington/go-libsass#pkg-constants
func SassOptionSetOutputStyle(goopts SassOptions, i int) {
	C.sass_option_set_output_style(goopts, uint32(i))
}

// SassOptionSetSourceComments toggles the output of line comments in CSS
func SassOptionSetSourceComments(goopts SassOptions, b bool) {
	C.sass_option_set_source_comments(goopts, C.bool(b))
}

// SassOptionSetOutputPath is used for output path.
// Output path is used for source map generating. LibSass will
// not write to this file, it is just used to create
// information in source-maps etc.
func SassOptionSetOutputPath(goopts SassOptions, path string) {
	C.sass_option_set_output_path(goopts, C.CString(path))
}

// SassOptionSetIncludePaths adds additional paths to look for input Sass
func SassOptionSetIncludePath(goopts SassOptions, path string) {
	C.sass_option_set_include_path(goopts, C.CString(path))
}

func SassOptionSetSourceMapEmbed(goopts SassOptions, b bool) {
	C.sass_option_set_source_map_embed(goopts, C.bool(b))
}

func SassOptionSetSourceMapContents(goopts SassOptions, b bool) {
	C.sass_option_set_source_map_contents(goopts, C.bool(b))
}

func SassOptionSetSourceMapFile(goopts SassOptions, path string) {
	C.sass_option_set_source_map_file(goopts, C.CString(path))
}

// SassOptionSetSourceMapRoot sets the sourceRoot in the resulting
// source map. sourceRoot is prepended to the sources referenced
// in the map. More Info: https://docs.google.com/document/d/1U1RGAehQwRypUTovF1KRlpiOFze0b-_2gc6fAH0KY0k/edit#heading=h.75yo6yoyk7x5
func SassOptionSetSourceMapRoot(goopts SassOptions, path string) {
	C.sass_option_set_source_map_root(goopts, C.CString(path))
}

func SassOptionSetOmitSourceMapURL(goopts SassOptions, b bool) {
	C.sass_option_set_omit_source_map_url(goopts, C.bool(b))
}

type SassImportEntry C.Sass_Import_Entry

// SassMakeImport creates an import for overriding behavior when
// certain imports are called.
func SassMakeImport(path string, base string, source string, srcmap string) SassImportEntry {
	impent := C.sass_make_import(C.CString(path), C.CString(base),
		C.CString(source), C.CString(srcmap))
	return (SassImportEntry)(impent)
}

type SassImporterFN C.Sass_Importer_Fn

func SassImporterGetFunction(goimp SassImporter) SassImporterFN {
	impfn := C.sass_importer_get_function(C.Sass_Importer_Entry(goimp))
	return (SassImporterFN)(impfn)
}

func SassImporterGetListEntry() {}

// SassMakeImport attaches a Go pointer to a Sass function. Go will
// eventually kill this behavior and we should find a new way to do this.
func SassMakeImporter(fn SassImporterFN, priority int, v interface{}) (SassImporter, error) {
	vv := reflect.ValueOf(v).Elem()
	if !vv.CanAddr() {
		return nil, errors.New("can not take address of")
	}
	// TODO: this will leak memory, the interface must be freed manually
	lst := C.sass_make_importer(C.Sass_Importer_Fn(fn), C.double(priority), unsafe.Pointer(vv.Addr().Pointer()))
	return (SassImporter)(lst), nil
}

// SassImporterSetListEntry updates an import list with a new entry. The
// index must exist in the importer list.
func SassImporterSetListEntry(golst SassImporterList, idx int, ent SassImporter) {
	C.sass_importer_set_list_entry(C.Sass_Importer_List(golst), C.size_t(idx), C.Sass_Importer_Entry(ent))
}

func SassOptionSetCImporters(goopts SassOptions, golst SassImporterList) {
	C.sass_option_set_c_importers(goopts, C.Sass_Importer_List(golst))
}

func SassOptionSetCFunctions() {

}
