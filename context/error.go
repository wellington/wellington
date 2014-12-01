package context

import "encoding/json"

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
func (ctx *Context) ProcessSassError(bs []byte) (string, error) {

	if len(bs) == 0 {
		return "", nil
	}

	err := json.Unmarshal(bs, &ctx.Errors)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

func (ctx *Context) Error() string {
	return ctx.errorString
}
