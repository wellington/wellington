package libsass

import (
	"io"

	"github.com/wellington/go-libsass/libs"
)

// ToScss converts Sass to Scss with libsass sass2scss.h
func ToScss(r io.Reader, w io.Writer) error {
	return libs.ToScss(r, w)
}
