package context

import (
	"encoding/json"
	"fmt"
)

type lError struct {
	Pos     int
	Message string
}

type lErrors struct {
	Errors []lError
	Pos    int
}

type SassError struct {
	Status, Line, Column int
	File, Message        string
}

// Error reads the original libsass error and creates helpful debuggin
// information for debuggin that error.
func (ctx *Context) ProcessSassError(bs []byte) error {

	if len(bs) == 0 {
		return nil
	}

	err := json.Unmarshal(bs, &ctx.Errors)
	if err != nil {
		return err
	}

	errors := ctx.Errors
	ctx.errorString = fmt.Sprintf("Error > %s:%d\n%s",
		errors.File, errors.Line, errors.Message)
	return nil
}

func (ctx *Context) Error() string {
	return ctx.errorString
}

// LineNumber attempts to resolve the file associated with
// a stdin:#
func (ctx *Context) ErrorLine() int {
	var n int
	fmt.Sscanf(ctx.Error(), "Error > stdin:%d", &n)
	return n
}
