package sprite_sass

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"regexp"
	"strings"
	"testing"
)

var rerandom *regexp.Regexp

var spritePreamble string

func init() {
	rerandom = regexp.MustCompile(`-\w{6}(?:\.(png|jpg))`)

	bs, err := ioutil.ReadFile("sass/_sprite.scss")
	if err != nil {
		log.Fatal(err)
	}
	spritePreamble = strings.TrimSuffix(string(bs), "\n")
}

func TestParserVar(t *testing.T) {
	p := Parser{}
	fread := fileReader("test/sass/_var.scss")
	bs, _ := p.Start(fread, "test/")
	output := string(bs)

	defer cleanUpSprites(p.Sprites)

	file, _ := ioutil.ReadFile("test/expected/_var.parser")
	e := string(file)
	if e != output {
		t.Errorf("File output did not match, \nexpected:\n%s\nwas:\n%s",
			e, output)
	}

}

func TestParserRelative(t *testing.T) {
	p := Parser{
		StaticDir: "test",
		BuildDir:  "test/build",
	}
	f, err := ioutil.ReadFile("sass/_sprite.scss")
	if err != nil {
		log.Fatal(err)
	}
	in := bytes.NewBuffer(f)
	in.WriteString(`div {
  background: image-url('img/139.png');
}`)
	e := fmt.Sprintf(`$rel: "..";
%s
div {
  background: image-url('img/139.png');
}`, spritePreamble)
	bs, _ := p.Start(in, "test")
	out := string(bs)
	defer cleanUpSprites(p.Sprites)
	if out != e {
		t.Errorf("Mismatch expected:\n%s\nwas:\n%s", e, out)
	}

}

func TestParserImporter(t *testing.T) {
	p := Parser{
		StaticDir: "test",
		BuildDir:  "test/build",
		Includes:  []string{"test/sass"},
	}

	bs, err := p.Start(fileReader("test/sass/import.scss"), "test/")
	if err != nil {
		log.Fatal(err)
	}
	output := string(bs)

	defer cleanUpSprites(p.Sprites)

	file, _ := ioutil.ReadFile("test/expected/import.parser")
	e := string(file)
	if e != output {
		t.Errorf("File output did not match, was:\n~%s~\nexpected:\n~%s~",
			output, e)
	}

	lines := map[int]string{
		0:  "../sass/sprite",
		36: "var",
		47: "string",
	}
	errors := false
	for i, v := range lines {
		if v != p.Line[i] {
			t.Errorf("Invalid expected: %s, was: %s", v, p.Line[i])
			errors = true
		}
	}
	if errors {
		fmt.Println(p.Line)
	}
}

func TestParseSpriteArgs(t *testing.T) {
	p := Parser{}
	in := bytes.NewBufferString(`$view_sprite: sprite-map("test/*.png",
  $normal-spacing: 2px,
  $normal-hover-spacing: 2px,
  $selected-spacing: 2px,
  $selected-hover-spacing: 2px);
  @include sprite-dimensions($view_sprite,140);
`)
	e := `$rel: ".";
$view_sprite: (); $view_sprite: map_merge($view_sprite,(139: (width: 96, height: 139, x: 0, y: 0, url: 'test-585dca.png'))); $view_sprite: map_merge($view_sprite,(140: (width: 96, height: 140, x: 0, y: 139, url: 'test-585dca.png'))); $view_sprite: map_merge($view_sprite,(pixel: (width: 1, height: 1, x: 0, y: 279, url: 'test-585dca.png')));
  @include sprite-dimensions($view_sprite,140);
`
	bs, _ := p.Start(in, "")
	out := string(bs)
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
	bs, _ := p.Start(r, "")
	res = string(bs)

	e = `$rel: ".";
p {
  $font-size: 12px;
  $line-height: 30px;
  font: #{$font-size}/#{$line-height};
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
		StaticDir: "test",
		GenImgDir: "test/build/img",
		BuildDir:  "test/build",
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
	defer cleanUpSprites(p.Sprites)

	if e := `$rel: "..";
$sprites: (); $sprites: map_merge($sprites,(139: (width: 96, height: 139, x: 0, y: 0, url: 'test/build/img/img-d65510.png'))); $sprites: map_merge($sprites,(140: (width: 96, height: 140, x: 0, y: 139, url: 'test/build/img/img-d65510.png')));
$sfile: sprite-file($sprites, 139);
div {
    height: image-height(sprite-file($sprites, 139));
    width: image-width(test/139.png);
    url: sprite-file($sprites, 139);
}`; e != out {
		t.Errorf("Mismatch expected:\n%s\nwas:\n%s\n", e, out)
	}
}

func TestParseImageUrl(t *testing.T) {
	return // Test no longer useful
	p := Parser{
		BuildDir: "/doop/doop",
	}
	in := bytes.NewBufferString(`background: image-url("test/140.png");`)
	var b bytes.Buffer
	//log.SetOutput(&b)
	bs, _ := p.Start(in, "")
	out := string(bs)

	if e := "can't make . relative to /doop/doop\n"; !strings.HasSuffix(
		b.String(), e) {
		t.Errorf("No error for bad relative path expected:\n%s\nwas:\n%s\n",
			e, b.String())
	}

	if e := "background: url(\"\");"; e != out {
		//t.Errorf("expected: %s, was: %s", e, out)
	}
	log.SetOutput(os.Stdout)
}
