package sprite_sass

import (
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

	file, _ := ioutil.ReadFile("test/sprite.parser")
	if strings.TrimSpace(string(file)) != strings.TrimSpace(output) {
		t.Errorf("File output did not match, was:\n%s\nexpected:\n%s", output, string(file))
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
