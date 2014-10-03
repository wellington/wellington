package sprite_sass

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"

	"regexp"
	"strings"
	"testing"
)

var rerandom *regexp.Regexp

func init() {
	rerandom = regexp.MustCompile(`-\w{6}(?:\.(png|jpg))`)
}

func TestParserVar(t *testing.T) {
	p := Parser{}
	fread := fileReader("test/_var.scss")
	output := string(p.Start(fread, "test/"))
	output = strings.TrimSpace(rerandom.ReplaceAllString(output, ""))
	defer cleanUpSprites(p.Sprites)
	file, _ := ioutil.ReadFile("test/var.parser")
	e := strings.TrimSpace(string(file))
	if e != output {
		t.Errorf("File output did not match, \nwas:\n%s\nexpected:\n%s",
			output, e)
	}

}

func TestParserImporter(t *testing.T) {
	p := Parser{}
	output := string(p.Start(fileReader("test/import.scss"), "test/"))
	output = strings.TrimSpace(rerandom.ReplaceAllString(output, ""))

	defer cleanUpSprites(p.Sprites)
	file, _ := ioutil.ReadFile("test/import.parser")
	e := strings.TrimSpace(string(file))
	if e != output {
		t.Errorf("File output did not match, was:\n%s\nexpected:\n%s",
			output, e)
	}
}

func TestParseSprite(t *testing.T) {
	p := Parser{}
	output := string(p.Start(fileReader("test/sprite.scss"), "test/"))
	output = rerandom.ReplaceAllString(output, "")

	defer cleanUpSprites(p.Sprites)
	file, _ := ioutil.ReadFile("test/sprite.parser")
	if strings.TrimSpace(string(file)) != strings.TrimSpace(output) {
		t.Errorf("File output did not match, was:\n%s\nexpected:\n%s", output, string(file))
	}
}

func TestParseSpriteArgs(t *testing.T) {
	p := Parser{}
	in := bytes.NewBufferString(`$view_sprite: sprite-map("test/*.png",
  $normal-spacing: 2px,
  $normal-hover-spacing: 2px,
  $selected-spacing: 2px,
  $selected-hover-spacing: 2px);
  @include sprite-dimensions($view_sprite,"140");
`)
	e := `
  width: 96px;
height: 140px;
`
	out := string(p.Start(in, ""))
	defer cleanUpSprites(p.Sprites)
	if out != e {
		t.Errorf("Mismatch expected:\n%s\nwas:\n%s", e, out)
	}
}

func TestParseInt(t *testing.T) {
	p := Parser{}
	var (
		e, res string
	)
	r := bytes.NewBufferString(`p {
	  $font-size: 12px;
	  $line-height: 30px;
	  font: #{$font-size}/#{$line-height};
	}`)

	res = string(p.Start(r, ""))

	e = `p {
	  $font-size: 12px;
	  $line-height: 30px;
	  font: 12px/30px;
	}`
	if e != res {
		t.Errorf("Mismatch expected:\n%s\nwas:\n%s", e, res)
	}
	p = Parser{}
	r = bytes.NewBufferString(`$name: foo;
$attr: border;
p.#{$name} {
  #{$attr}-color: blue;
}`)
	res = string(p.Start(r, ""))

	e = `$name: foo;
$attr: border;
p.foo {
  border-color: blue;
}`
	if e != res {
		t.Errorf("Mismatch expected:\n%s\nwas:\n%s", e, res)
	}
}

func TestParseComment(t *testing.T) {
	p := Parser{}
	res := string(p.Start(fileReader("test/_comment.scss"), "test/"))
	res = strings.TrimSpace(rerandom.ReplaceAllString(res, ""))
	e := strings.TrimSpace(fileString("test/comment.parser"))

	if res != e {
		t.Errorf("Comment parsing failed was:"+
			"%s\n exp:%s\n", res, e)
	}
}

func TestParseMixin(t *testing.T) {
	p := Parser{}
	res := string(p.Start(fileReader("test/mixin.scss"), ""))
	e := fileString("test/mixin.parser")

	if res != e {
		t.Errorf("Mixin parsing failed\n  was:%s\n expected:%s\n",
			res, e)
	}
}

func TestParseImage(t *testing.T) {
	p := Parser{}
	in := bytes.NewBufferString(`$sprites: sprite-map("test/*.png");
$sfile: sprite-file($sprites, 139);
div {
    height: image-height(sprite-file($sprites, 139));
    width: image-width(test/139.png);
    url: sprite-file($sprites, 139);
}`)

	out := string(p.Start(in, ""))
	defer cleanUpSprites(p.Sprites)

	if e := `

div {
    height: 139px;
    width: 96px;
    url: test/139.png;
}`; e != out {
		t.Errorf("Mismatch expected:\n%s\nwas:\n%s\n", e, out)
	}
}

func TestParseImageUrl(t *testing.T) {
	p := Parser{
		BuildDir: "/doop/doop",
	}
	in := bytes.NewBufferString(`background: image-url("test/140.png");`)
	var b bytes.Buffer
	log.SetOutput(&b)
	out := string(p.Start(in, ""))

	if e := "can't make . relative to /doop/doop\n"; !strings.HasSuffix(
		b.String(), e) {
		t.Errorf("No error for relative expected: %s, was: %s",
			e, b.String())
	}

	if e := "background: url(\"\");"; e != out {
		t.Errorf("expected: %s, was: %s", e, out)
	}
	log.SetOutput(os.Stdout)
}
