package sprite_sass

import (
	"bytes"
	"io/ioutil"

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
