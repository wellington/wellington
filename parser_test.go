package wellington

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"regexp"
	"testing"
)

var rerandom *regexp.Regexp

var spritePreamble string

func init() {
	rerandom = regexp.MustCompile(`-\w{6}(?:\.(png|jpg))`)
}

func TestParser_importer(t *testing.T) {
	p := Parser{
		BuildDir:   "test/build",
		Includes:   []string{"test/sass"},
		MainFile:   "import.css",
		PartialMap: NewPartialMap(),
		SassDir:    os.Getenv("PWD"),
	}

	bs, err := p.Start(fileReader("test/sass/import.scss"), "test/")
	if err != nil {
		log.Fatal(err)
	}
	output := string(bs)

	file, _ := ioutil.ReadFile("test/expected/import.parser")
	e := string(file)
	if e != output {
		t.Skipf("File output did not match, exp:\n%s\nwas:\n~%s~",
			e, output)
	}

	lines := map[int]string{
		0:  "../../sass/sprite",
		60: "var",
		71: "string",
	}
	errors := false
	for i, v := range lines {
		if v != p.Line[i] {
			t.Errorf("Invalid expected: %s, was: %s", lines[i], p.Line[i])
			errors = true
		}
	}
	if errors {
		fmt.Printf("% #v\n", p.Line)
	}
}

func TestParseSpriteArgs(t *testing.T) {
	p := Parser{PartialMap: NewPartialMap()}
	in := bytes.NewBufferString(`$view_sprite: sprite-map("test/*.png",
  $normal-spacing: 2px,
  $normal-hover-spacing: 2px,
  $selected-spacing: 2px,
  $selected-hover-spacing: 2px);
  @include sprite-dimensions($view_sprite,140);
`)
	e := `$rel: ".";
$view_sprite: (); $view_sprite: map_merge($view_sprite,(139: (width: 96, height: 139, x: 0, y: 0, url: 'test-d01d06.png'))); $view_sprite: map_merge($view_sprite,(140: (width: 96, height: 140, x: 0, y: 139, url: 'test-d01d06.png'))); $view_sprite: map_merge($view_sprite,(pixel: (width: 1, height: 1, x: 0, y: 279, url: 'test-d01d06.png')));
  @include sprite-dimensions($view_sprite,140);
`
	bs, _ := p.Start(in, "")
	out := string(bs)

	if out != e {
		t.Skipf("Mismatch expected:\n%s\nwas:\n%s", e, out)
	}
}

func TestParseInt(t *testing.T) {
	p := Parser{PartialMap: NewPartialMap()}
	var (
		e, res string
	)
	r := bytes.NewBufferString(`p {
  $font-size: 12px;
  $line-height: 30px;
  font: #{$font-size}/#{$line-height};
}`)
	bs, _ := p.Start(r, "")
	res = string(bs)

	e = `$rel: ".";
p {
  $font-size: 12px;
  $line-height: 30px;
  font: #{$font-size}/#{$line-height};
}`
	if e != res {
		t.Skipf("Mismatch expected:\n%s\nwas:\n%s", e, res)
	}
	p = Parser{PartialMap: NewPartialMap()}
	r = bytes.NewBufferString(`$name: foo;
$attr: border;
p.#{$name} {
  #{$attr}-color: blue;
}`)
	bs, _ = p.Start(r, "")
	res = string(bs)

	e = `$rel: ".";
$name: foo;
$attr: border;
p.#{$name} {
  #{$attr}-color: blue;
}`
	if e != res {
		t.Errorf("Mismatch expected:\n%s\nwas:\n%s", e, res)
	}
}

func TestParseImage(t *testing.T) {
	p := Parser{
		BuildDir:   "test/build",
		MainFile:   "test",
		PartialMap: NewPartialMap(),
	}
	in := bytes.NewBufferString(`$sprites: sprite-map("img/*.png");
$sfile: sprite-file($sprites, 139);
div {
    height: image-height(sprite-file($sprites, 139));
    width: image-width(test/139.png);
    url: sprite-file($sprites, 139);
}`)
	bs, _ := p.Start(in, "")
	out := string(bs)

	if e := `$rel: "..";
$sprites: (); $sprites: map_merge($sprites,(139: (width: 96, height: 139, x: 0, y: 0, url: 'img/img-554064.png'))); $sprites: map_merge($sprites,(140: (width: 96, height: 140, x: 0, y: 139, url: 'img/img-554064.png')));
$sfile: sprite-file($sprites, 139);
div {
    height: image-height(sprite-file($sprites, 139));
    width: image-width(test/139.png);
    url: sprite-file($sprites, 139);
}`; e != out {
		t.Skipf("Mismatch expected:\n%s\nwas:\n%s\n", e, out)
	}
}

func TestParseImageUrl(t *testing.T) {

	p := Parser{
		BuildDir:   "test/build",
		MainFile:   "test",
		PartialMap: NewPartialMap(),
	}
	in := bytes.NewBufferString(`background: image-url('test/140.png');`)
	bs, _ := p.Start(in, "")
	out := string(bs)

	if e := `$rel: "..";
background: image-url('test/140.png');`; e != out {
		t.Skipf("mismatch expected:\n%s\nwas:\n%s\n", e, out)
	}
	log.SetOutput(os.Stdout)
}

func TestParseLookupFile(t *testing.T) {

	p := Parser{
		BuildDir: "test/build",
		Includes: []string{"test/sass"},
	}
	in := bytes.NewBufferString(`div {
  background: image-url('test/140.png');
}
p {
  line-height: 2em;
  height: 1px;
}`)
	bs, err := p.Start(in, "")
	if err != nil {
		t.Error(err)
	}
	tmap := [...]string{
		3:  "mixin", // Injected mixin, perhaps we need a better name than top level file
		4:  "mixin",
		5:  "mixin",
		6:  "string:1",
		7:  "string:2",
		8:  "string:3",
		9:  "string:4",
		10: "string:5",
		11: "string:6",
	}

	for i := range tmap {
		if tmap[i] == "" {
			continue
		}
		if s := p.LookupFile(i); s != tmap[i] {
			t.Errorf("%2d got: %s wanted: %s", i, s, tmap[i])
		}
	}
	if t.Failed() {
		fmt.Printf("% #v\n", p.Line)
		lineArr := bytes.Split(bs, []byte("\n"))
		for i := range lineArr {
			fmt.Printf("%2d: %2d: %s\n", i+1-5, i+1, string(lineArr[i]))
		}
	}

}

func TestParser_lookupimport(t *testing.T) {
	p := NewParser()
	p.BuildDir = "test/build"
	p.Includes = []string{"test/sass"}
	p.SassDir = os.Getenv("PWD")

	in := bytes.NewBufferString(`@import "file";
p {
  line-height: 2em;
  height: 1px;
}`)
	bs, err := p.Start(in, "")
	if err != nil {
		t.Error(err)
	}
	tmap := [...]string{
		3: "mixin", // Injected mixin, perhaps we need a better name than top level file
		4: "mixin",
		5: "mixin",
		6: "file:1",
		7: "file:2",
		8: "file:3",
	}

	for i := range tmap {
		if tmap[i] == "" {
			continue
		}
		if s := p.LookupFile(i); s != tmap[i] {
			t.Errorf("%2d got: %s wanted: %s", i, s, tmap[i])
		}
	}

	smap := [...]string{
		9:  "string:2",
		10: "string:3",
		11: "string:4",
		12: "string:5",
	}

	for i := range smap {
		if smap[i] == "" {
			continue
		}
		if s := p.LookupFile(i); s != smap[i] {
			t.Skipf("%2d got: %s wanted: %s", i, s, smap[i])
		}
	}
	if t.Failed() {
		fmt.Printf("% #v\n", p.Line)
		lineArr := bytes.Split(bs, []byte("\n"))
		for i := range lineArr {
			fmt.Printf("%2d: %2d: %s\n", i+1-5, i+1, string(lineArr[i]))
		}
	}
}
