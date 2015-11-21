package spritewell

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/url"
	"unicode/utf8"
)

// IsSVG attempts to determine if a reader contains an SVG
func IsSVG(r io.Reader) bool {

	// Copy first 1k block and look for <svg
	var buf bytes.Buffer
	io.CopyN(&buf, r, bytes.MinRead)

	s := bufio.NewScanner(&buf)
	s.Split(bufio.ScanWords)
	mat := []byte("<svg")
	for s.Scan() {
		if bytes.Equal(mat, s.Bytes()) {
			return true
		}
		// Guesstimate that SVG with non-utf8 is no SVG at all
		if !utf8.Valid(s.Bytes()) {
			return false
		}
	}
	return false
}

// Inline accepts an io.Reader and writes to io.Writer.  Binary
// data is base64 encoded, but base64 encoding is optional for
// svg.
func Inline(r io.Reader, w io.Writer, encode ...bool) error {

	// Check if SVG
	var buf bytes.Buffer
	tr := io.TeeReader(r, &buf)
	if IsSVG(tr) {
		enc := len(encode) > 0 && encode[0]
		mr := io.MultiReader(&buf, r)
		inlineSVG(w, mr, enc)
		return nil
	}
	mr := io.MultiReader(&buf, r)
	m, _, err := image.Decode(mr)
	if err != nil {
		return err
	}
	w.Write([]byte(`url("data:image/png;base64,`))
	bw := base64.NewEncoder(base64.StdEncoding, w)
	err = png.Encode(bw, m)
	w.Write([]byte(`")`))
	return err
}

// inlinesvg returns a byte slice that is utf8 compliant or base64
// encoded
func inlineSVG(w io.Writer, r io.Reader, encode bool) {
	if encode {
		bw := base64.NewEncoder(base64.StdEncoding, w)
		w.Write([]byte(`url("data:image/svg+xml;base64,`))
		io.Copy(bw, r)
		w.Write([]byte(`")`))
		return
	}

	w.Write([]byte(`url("data:image/svg+xml;utf8,`))
	// Exhaust the buffer and do some sloppy regex stuff to the input
	// TODO: convert url encoding to streaming reader/writer
	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Strip unnecessary newlines
	input := bytes.Replace(buf.Bytes(), []byte("\r\n"), []byte(""), -1)
	// input = bytes.Replace(input, []byte(`"`), []byte("'"), -1)
	// reg := regexp.MustCompile(`>\\s+<`)
	// input = reg.ReplaceAll(input, []byte("><"))

	// url.String() properly escapes paths
	u := &url.URL{Path: string(input)}
	io.WriteString(w, u.String())
	w.Write([]byte(`")`))
}
