package libsass

import (
	"fmt"
	"testing"
)

func init() {
	fmt.Println(LIBSASS_VERSION)
}

func TestVersion(t *testing.T) {
	t.Error(LIBSASS_VERSION)
}
