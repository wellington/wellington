package sprite_sass_test

import (
	"io/ioutil"
	"regexp"
	"strings"
	"testing"

	. "github.com/drewwells/sprite_sass"
)

var rerandom *regexp.Regexp

func init() {
	rerandom = regexp.MustCompile(`-\w{6}(?:\.png)`)
}

func TestParserVar(t *testing.T) {
	p := Parser{}
	output := string(p.Start("test/_var.scss"))
	output = rerandom.ReplaceAllString(output, "")

	file, _ := ioutil.ReadFile("test/var.parser")
	if strings.TrimSpace(string(file)) != strings.TrimSpace(output) {
		t.Errorf("File output did not match, was:\n%s\nexpected:\n%s", output, string(file))
	}
}

func TestParserImporter(t *testing.T) {

	p := Parser{}
	output := string(p.Start("test/import.scss"))
	output = rerandom.ReplaceAllString(output, "")

	file, _ := ioutil.ReadFile("test/import.parser")
	if strings.TrimSpace(string(file)) != strings.TrimSpace(output) {
		t.Errorf("File output did not match, was:\n%s\nexpected:\n%s", output, string(file))
	}
}

func TestParseComment(t *testing.T) {
	p := Parser{}

	res := string(p.Start("test/_comment.scss"))
	res = strings.TrimSpace(rerandom.ReplaceAllString(res, ""))
	e := strings.TrimSpace(fileString("test/comment.parser"))

	if res != e {
		t.Errorf("Comment parsing failed was:"+
			"%s\n exp:%s\n", res, e)
	}
}
