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
	ctx.AddImport("a", "a { color: blue; }")
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
	ctx.AddImport("a", "a { color: blue; }")
	ctx.AddImport("b", "b { font-weight: bold; }")
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

	ctx.AddImport("branch", `%branch { color: brown; }`)
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

	ctx.AddImport("branch", `@import "leaf";
%branch { color: brown; }`)
	ctx.AddImport("leaf", "%leaf { color: green; }")
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
	t.Skip("Skip this test for now")
	in := bytes.NewBufferString(`@import "branch";
div.branch {
  @extend %branch;
  div.leaf {
    @extend %leaf;
  }
}`)

	var out bytes.Buffer
	ctx := Context{}
	ctx.AddImport("nope", `@import "leaf";
%branch { color: brown; }`)
	err := ctx.Compile(in, &out)

	t.Error(err)
}
