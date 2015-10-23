package wellington

import "testing"

func TestPath_recurse(t *testing.T) {

	paths := pathsToFiles([]string{"test/includes"}, true)
	if len(paths) != 1 {
		t.Fatal("wrong number of returned paths")
	}

	// This is going to be a really annoying test, should setup a special
	// directory to test this.
	paths = pathsToFiles([]string{"test"}, true)
	if e := 18; len(paths) != e {
		t.Errorf("got: %d wanted: %d", len(paths), e)
	}
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

func TestRelative(t *testing.T) {
	paths := []string{"../test/sass", "test/sass", "/tmp"}
	var r string
	r = relative(paths, "test/sass/file.scss")
	if e := `.`; r != e {
		t.Errorf("got: %s wanted: %s", r, e)
	}

	r = relative(paths, "test/sass/subdir/file.scss")
	if e := `subdir`; r != e {
		t.Errorf("got: %s wanted: %s", r, e)
	}

	r = relative(paths, "../test/sass/subdir/file.scss")
	if e := `subdir`; r != e {
		t.Errorf("got: %s wanted: %s", r, e)
	}

	r = relative(paths, "/tmp/sass/file.scss")
	if e := `sass`; r != e {
		t.Errorf("got: %s wanted: %s", r, e)
	}

}
