package libsass

import (
	"errors"
	"testing"
)

func TestRegisterHandler(t *testing.T) {
	l := len(globalHandlers)
	RegisterHandler("foo",
		func(v interface{}, csv SassValue, rsv *SassValue) error {
			u, _ := Marshal(false)
			*rsv = u
			return nil
		})
	if e := l + 1; len(globalHandlers) != e {
		t.Errorf("got: %d wanted: %d", len(globalHandlers), e)
	}
}

func TestError_simple(t *testing.T) {
	err := errors.New("help me")
	sv := Error(err)

	var s string
	Unmarshal(sv, &s)
	if err.Error() != s {
		t.Errorf("got: %s wanted: %s", s, err)
	}
}
