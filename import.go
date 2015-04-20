package wellington

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"

	libsass "github.com/wellington/libsass"
)

type nopCloser struct {
	io.Reader
}

func (n nopCloser) Close() error { return nil }

// readSass retrives a file from path. If found, it converts Sass
// to Scss or returns found Scss;
func readSass(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return ToScssReader(file)
}

// ToScssReader ...
func ToScssReader(r io.Reader) (io.ReadCloser, error) {
	var (
		buf bytes.Buffer
	)

	tr := io.TeeReader(r, &buf)
	if IsSass(tr) {

		var ibuf bytes.Buffer
		libsass.ToScss(io.MultiReader(&buf, r), &ibuf)
		return nopCloser{&ibuf}, nil
	}
	mr := io.MultiReader(&buf, r)

	return nopCloser{mr}, nil
}

// IsSass determines if the given reader is Sass (not Scss).
// This is predicted by the presence of semicolons
func IsSass(r io.Reader) bool {
	scanner := bufio.NewScanner(r)
	var cmt bool
	var last string
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		last = text
		// FIXME: This is not a suitable way to detect comments
		if strings.HasPrefix(text, "/*") {
			cmt = true
		}
		if strings.HasSuffix(text, "*/") {
			cmt = false
			continue
		}
		if cmt {
			continue
		}
		if strings.HasSuffix(text, "{") ||
			strings.HasSuffix(text, "}") {
			return false
		}
		if strings.HasSuffix(text, ";") {
			return false
		}
	}
	// If type is still undecided and file ends with comment, assume
	// this is a scss file
	if strings.HasSuffix(last, "*/") {
		return false
	}
	return true
}
