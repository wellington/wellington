package context

// #include <stdlib.h>
// #include "sass2scss.h"
import "C"

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"unsafe"
)

// ToScss converts Sass to Scss with libsass sass2scss.h
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

func testToScss(t *testing.T) {
	file, err := os.Open("../test/whitespace/one.sass")
	if err != nil {
		t.Fatal(err)
	}
	e := `$font-stack:    Helvetica, sans-serif;
$primary-color: #333;

body {
  font: 100% $font-stack;
  color: $primary-color; }
`

	var bytes bytes.Buffer
	ToScss(file, &bytes)

	if bytes.String() != e {
		t.Errorf("got:\n%s\nwanted:\n%s", bytes.String(), e)
	}
}
