package wellington

import "testing"

func TestProcessSass(t *testing.T) {

	p := NewParser()
	p.Includes = []string{"test/whitespace"}

	in := fileReader("test/whitespace/import.sass")

	bs, err := p.Start(in, "")
	if err != nil {
		t.Fatal(err)
	}

	e := `@mixin sprite-dimensions($map, $name) {
  $file: sprite-file($map, $name);
  height: image-height($file);
  width: image-width($file);
}

$font-stack:    Helvetica, sans-serif;
$primary-color: #333;


body {
  font: 100% $font-stack;
  background-color: $primary-color; }
`

	if e != string(bs) {
		t.Errorf("got:\n%s\nwanted:\n%s", string(bs), e)
	}
}
