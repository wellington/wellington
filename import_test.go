package wellington

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestReadSass(t *testing.T) {
	e := []byte(`
html,
body,
ul,
ol {
  margin:  0;
  padding: 0; }
`)

	r, err := readSass("test/whitespace/base.sass")
	if err != nil {
		t.Fatal(err)
	}

	bs, _ := ioutil.ReadAll(r)
	if bytes.Compare(bs, e) != 0 {
		t.Fatalf("got: %s\n wanted: %s", bs, e)
	}
}

func TestIsSass(t *testing.T) {
	buf := bytes.NewBufferString(`/*
Big stupid comment
*/`)
	yes := IsSass(buf)

	if yes {
		t.Fatal("This is not sass")
	}
}
