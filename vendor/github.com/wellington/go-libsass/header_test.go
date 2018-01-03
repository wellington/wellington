package libsass

import (
	"bytes"
	"testing"
)

func TestSassHeader_single(t *testing.T) {
	in := bytes.NewBufferString(`div {
  @include mix();
}
`)

	var out bytes.Buffer
	ctx := newContext()
	ctx.Headers.Add(`@mixin mix() {
  width: 50px;
}
`)

	err := ctx.compile(&out, in)
	if err != nil {
		for _, h := range ctx.Headers.h {
			t.Logf("% #v\n", h)
		}
		t.Fatal(err)
	}
	e := `div {
  width: 50px; }
`
	if e != out.String() {
		t.Fatalf("got:\n%s\nwanted:\n%s", out.String(), e)
	}

}

func TestSassHeader_multi(t *testing.T) {
	in := bytes.NewBufferString(`div {
  @include mix();
  color: red();
}
`)

	var out bytes.Buffer
	ctx := newContext()
	ctx.Headers.Add(`@mixin mix() {
  width: 50px;
}
`)
	ctx.Headers.Add(`@function red() { @return red; }`)

	err := ctx.compile(&out, in)
	if err != nil {
		t.Fatal(err)
	}
	e := `div {
  width: 50px;
  color: red; }
`
	if e != out.String() {
		t.Fatalf("got:\n%s\nwanted:\n%s", out.String(), e)
	}

}
