package libsass

import (
	"bytes"
	"testing"
)

func TestToScss(t *testing.T) {
	in := bytes.NewBufferString(`html,
body,
ul,
ol
  margin:  0
  padding: 0
`)

	var b bytes.Buffer
	err := ToScss(in, &b)
	if err != nil {
		t.Fatal(err)
	}

	e := `html,
body,
ul,
ol {
  margin:  0;
  padding: 0; }
`
	if b.String() != e {
		t.Errorf("got:\n%s\nwanted:\n%s", b.String(), e)
	}

}
