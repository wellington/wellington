package context

import (
	"bytes"
	"os"
	"testing"
)

func TestToScss(t *testing.T) {
	t.Skip("disabled after extraction")
	file, err := os.Open("../test/whitespace/one.sass")
	if err != nil {
		t.Fatal(err)
	}
	e := `$font-stack:    Helvetica, sans-serif;
$primary-color: #333;
`
	var b bytes.Buffer
	ToScss(file, &b)

	if b.String() != e {
		t.Errorf("got:\n%s\nwanted:\n%s", b.String(), e)
	}

	s := []byte(`=border-radius($radius)
  -webkit-border-radius: $radius
  -moz-border-radius:    $radius
  -ms-border-radius:     $radius
  border-radius:         $radius

.box
  +border-radius(10px)`)

	var in bytes.Buffer
	b.Reset()
	in.Write(s)

	ToScss(&in, &b)

	e = `@mixin border-radius($radius) {
  -webkit-border-radius: $radius;
  -moz-border-radius:    $radius;
  -ms-border-radius:     $radius;
  border-radius:         $radius; }

.box {
  @include border-radius(10px); }
`

	if b.String() != e {
		t.Errorf("got:\n%s\nwanted:\n%s", b.String(), e)
	}
}
