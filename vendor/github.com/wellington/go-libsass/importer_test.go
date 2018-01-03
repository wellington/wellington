package libsass

import (
	"bytes"
	"testing"
)

func TestSassImport_single(t *testing.T) {
	in := bytes.NewBufferString(`@import "a";`)

	var out bytes.Buffer
	ctx := newContext()
	ctx.Imports.m = make(map[string]Import)
	ctx.Imports.Add("", "a", []byte("a { color: blue; }"))
	err := ctx.compile(&out, in)
	if err != nil {
		t.Fatal(err)
	}
	e := `a {
  color: blue; }
`
	if e != out.String() {
		t.Fatalf("got:\n%s\nwanted:\n%s", out.String(), e)
	}

}

func TestSassImport_file(t *testing.T) {

	var out bytes.Buffer
	ctx := newContext()
	ctx.Imports.m = make(map[string]Import)
	ctx.Imports.Add("test/scss/file.scss", "a", []byte("a { color: blue; }"))
	err := ctx.fileCompile("test/scss/file.scss", &out, "", "")
	if err != nil {
		t.Fatal(err)
	}
	e := `a {
  color: blue; }
`

	if e != out.String() {
		t.Fatalf("got:\n%s\nwanted:\n%s", out.String(), e)
	}

}

func TestSassImport_multi(t *testing.T) {
	in := bytes.NewBufferString(`@import "a";
@import "b";`)

	var out bytes.Buffer
	ctx := newContext()
	ctx.Imports.m = make(map[string]Import)
	ctx.Imports.Add("", "a", []byte("a { color: blue; }"))
	ctx.Imports.Add("", "b", []byte("b { font-weight: bold; }"))
	err := ctx.compile(&out, in)
	if err != nil {
		t.Fatal(err)
	}
	e := `a {
  color: blue; }

b {
  font-weight: bold; }
`

	if e != out.String() {
		t.Fatalf("got:\n%s\nwanted:\n%s", out.String(), e)
	}

}

func TestSassImporter_placeholder(t *testing.T) {
	in := bytes.NewBufferString(`@import "branch";
div.branch {
  @extend %branch;
}`)

	var out bytes.Buffer
	ctx := newContext()
	ctx.Imports.m = make(map[string]Import)
	ctx.Imports.Add("", "branch", []byte(`%branch { color: brown; }`))
	err := ctx.compile(&out, in)
	if err != nil {
		t.Fatal(err)
	}
	e := `div.branch {
  color: brown; }
`
	if e != out.String() {
		t.Fatalf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestSassImporter_nested_placeholder(t *testing.T) {
	in := bytes.NewBufferString(`@import "branch";
div.branch {
  @extend %branch;
  div.leaf {
    @extend %leaf;
  }
}`)

	var out bytes.Buffer
	ctx := newContext()
	ctx.Imports.m = make(map[string]Import)
	ctx.Imports.Add("", "branch", []byte(`@import "leaf";
%branch { color: brown; }`))
	ctx.Imports.Add("branch", "leaf", []byte("%leaf { color: green; }"))
	err := ctx.compile(&out, in)
	if err != nil {
		t.Fatal(err)
	}
	e := `div.branch div.leaf {
  color: green; }

div.branch {
  color: brown; }
`
	if e != out.String() {
		t.Fatalf("got:\n%s\nwanted:\n%s", out.String(), e)
	}
}

func TestSassImporter_invalidimport(t *testing.T) {
	in := bytes.NewBufferString(`@import "branch";
div.branch {
  @extend %branch;
  div.leaf {
    @extend %leaf;
  }
}`)

	var out bytes.Buffer
	ctx := newContext()
	ctx.Imports.m = make(map[string]Import)
	err := ctx.compile(&out, in)
	if err == nil {
		t.Fatal("No error thrown for missing import")
	}
	e := `Error > stdin:1
File to import not found or unreadable: branch
Parent style sheet: stdin
@import "branch";
div.branch {
  @extend %branch;
  div.leaf {
    @extend %leaf;
  }
}
`
	if e != err.Error() {
		t.Fatalf("got:\n%s\nwanted:\n%s",
			err.Error(), e)
	}

}

func TestSassImporter_notfound(t *testing.T) {

	in := bytes.NewBufferString(`@import "branch";
div.branch {
  @extend %branch;
  div.leaf {
    @extend %leaf;
  }
}`)

	var out bytes.Buffer
	ctx := newContext()
	ctx.Imports.m = make(map[string]Import)
	ctx.Imports.Add("", "nope", []byte(`@import "leaf";
%branch { color: brown; }`))
	err := ctx.compile(&out, in)

	e := `Error > stdin:1
File to import not found or unreadable: branch
Parent style sheet: stdin
@import "branch";
div.branch {
  @extend %branch;
  div.leaf {
    @extend %leaf;
  }
}
`
	if e != err.Error() {
		t.Errorf("got:\n%s\nwant:\n%s", err, e)
	}

}
