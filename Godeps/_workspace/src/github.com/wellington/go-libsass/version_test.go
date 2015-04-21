package context

import "testing"

func TestVersion(t *testing.T) {
	if len(Version()) == 0 {
		t.Fatal("No version reported")
	}
}
