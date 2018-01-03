package libsass

import "testing"

func TestVersion(t *testing.T) {
	if len(Version()) == 0 {
		t.Fatal("No version reported")
	}
	if Version() == "[NA]" {
		t.Fatalf("got: %s wanted: versioned lib", Version())
	}
}
