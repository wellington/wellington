package libs

// #include <stdlib.h>
// #include "sass2scss.h"
import "C"
import (
	"io"
	"io/ioutil"
	"unsafe"
)

// ToScss converts Sass to Scss using sass2scss. Readers and Writers are
// used, but libsass does not support streaming.
func ToScss(r io.Reader, w io.Writer) error {
	bs, _ := ioutil.ReadAll(r)
	in := C.CString(string(bs))
	defer C.free(unsafe.Pointer(in))

	chars := C.sass2scss(
		// FIXME: readers would be much more efficient
		in,
		// SASS2SCSS_PRETTIFY_1 Egyptian brackets
		C.int(1),
	)
	_, err := io.WriteString(w, C.GoString(chars))
	return err
}
