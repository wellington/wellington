package wellington

import (
	"path/filepath"
	"testing"
)

func TestPartialMap(t *testing.T) {
	path, _ := filepath.Abs("test/sass/import.scss")
	p := Parser{
		BuildDir:   "test/build",
		Includes:   []string{"test/sass"},
		MainFile:   path,
		PartialMap: NewPartialMap(),
	}

	p.Start(fileReader("test/sass/import.scss"), "test/")
	if len(p.PartialMap.M) != 2 {
		t.Errorf("SafePartialMap expected Size: 1 got: %d", len(p.PartialMap.M))
	}

	for k := range p.PartialMap.M {
		if len(p.PartialMap.M[k]) != 1 {
			t.Errorf("SafePartialMap for test/sass/import.scss expected 1 top level file, got %d", len(p.PartialMap.M[k]))
		}
		if p.PartialMap.M[k][0] != path {
			t.Errorf("SafePartialMap expected all top level files to be %s, got %s", path, p.PartialMap.M[k][0])
		}
	}
}
