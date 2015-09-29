package wellington

import (
	"fmt"
	"testing"
)

func TestPath_recurse(t *testing.T) {

	paths := pathsToFiles([]string{"test/includes"}, true)
	if len(paths) != 1 {
		t.Fatal("wrong number of returned paths")
	}

	// This is going to be a really annoying test, should setup a special
	// directory to test this.
	paths = pathsToFiles([]string{"test"}, true)
	if e := 15; len(paths) != e {
		t.Errorf("got: %d wanted: %d", len(paths), e)
	}
	fmt.Println(paths)
}

func TestPath_files(t *testing.T) {

	paths := pathsToFiles([]string{"test/includes"}, false)
	if len(paths) != 1 {
		t.Fatal("wrong number of returned paths")
	}

	// This is going to be a really annoying test, should setup a special
	// directory to test this.
	paths = pathsToFiles([]string{"test"}, false)
	if e := 2; len(paths) != e {
		t.Errorf("got: %d wanted: %d", len(paths), e)
	}
}
