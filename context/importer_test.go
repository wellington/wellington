package context

import (
	"bytes"
	"strings"
	"testing"
)

func TestSassImport_single(t *testing.T) {
	in := bytes.NewBufferString(`@import "a";`)

	var out bytes.Buffer
	ctx := Context{}
	ctx.Imports.m = make(map[string]Import)
	ctx.Imports.Add("a", []byte("a { color: blue; }"))
	err := ctx.Compile(in, &out)
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
	ctx := Context{}
	ctx.Imports.m = make(map[string]Import)
	ctx.Imports.Add("a", []byte("a { color: blue; }"))
	ctx.Imports.Add("b", []byte("b { font-weight: bold; }"))
	err := ctx.Compile(in, &out)
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
	ctx := Context{}
	ctx.Imports.m = make(map[string]Import)
	ctx.Imports.Add("branch", []byte(`%branch { color: brown; }`))
	err := ctx.Compile(in, &out)
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
	ctx := Context{}
	ctx.Imports.m = make(map[string]Import)
	ctx.Imports.Add("branch", []byte(`@import "leaf";
%branch { color: brown; }`))
	ctx.Imports.Add("leaf", []byte("%leaf { color: green; }"))
	err := ctx.Compile(in, &out)
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
	ctx := Context{}
	ctx.Imports.m = make(map[string]Import)
	err := ctx.Compile(in, &out)
	if err == nil {
		t.Fatal("No error thrown for missing import")
	}
	e := `Error > stdin:1
file to import not found or unreadable: branch
Current dir: ` + `
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
			strings.Replace(err.Error(), " ", "*", -1),
			strings.Replace(e, " ", "*", -1))
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
	ctx := Context{}
	ctx.Imports.m = make(map[string]Import)
	ctx.Imports.Add("nope", []byte(`@import "leaf";
%branch { color: brown; }`))
	err := ctx.Compile(in, &out)

	e := `Error > stdin:1
file to import not found or unreadable: branch
Current dir: ` + `
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
