package libsass

import (
	"encoding/json"
	"fmt"
)

// SassError represents an error object returned from Sass.  SassError
// stores useful information for bubbling up libsass errors.
type SassError struct {
	Status, Line, Column int
	File, Message        string
}

// ProcessSassError reads the original libsass error and creates helpful debuggin
// information for debuggin that error.
func (ctx *compctx) ProcessSassError(bs []byte) error {

	if len(bs) == 0 {
		return nil
	}

	err := json.Unmarshal(bs, &ctx.err)
	if err != nil {
		return err
	}

	errors := ctx.err
	ctx.errorString = fmt.Sprintf("Error > %s:%d\n%s",
		errors.File, errors.Line, errors.Message)
	return nil
}

func (ctx *compctx) Error() string {
	return ctx.errorString
}

// Reset returns removes all error state information.
func (ctx *compctx) Reset() {
	ctx.errorString = ""
}

// ErrorLine attempts to resolve the file associated with
// a stdin:#
func (ctx *compctx) ErrorLine() int {
	var n int
	fmt.Sscanf(ctx.Error(), "Error > stdin:%d", &n)
	return n
}
