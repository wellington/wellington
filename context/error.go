package context

import (
	"bufio"
	"encoding/json"
	"log"
	"strings"
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

func (ctx *Context) ErrorTokenizer(e SassError) lErrors {
	errors := []lError{}
	r := strings.NewReplacer(":", " ", ",", " ")
	src := e.Message
	line := e.Line - 1
	src = r.Replace(src)
	scanner := bufio.NewScanner(strings.NewReader(src))
	scanner.Split(bufio.ScanWords)
	var (
		str string
	)
	for scanner.Scan() {

		if scanner.Text() != "Backtrace" && scanner.Text() != "stdin" {
			str += scanner.Text() + " "
		} else {
			// Disable backtrace parsing until new format is sorted out
			break
			if line == 0 && str == "" {
			} else {
				le := lError{
					Pos:     line,
					Message: strings.TrimSpace(str)}
				errors = append(errors, le)
				str = ""
				line = 0
			}
		}
	}

	// This looks a little stupid, perhaps simplify the nested
	// errors another way
	errors = append(errors, lError{
		Pos:     line,
		Message: strings.TrimSpace(str)})

	ctx.errors = lErrors{
		Pos:    line,
		Errors: errors,
	}
	return ctx.errors
}

/*
{
	"status": 1,
	"file": "stdin",
	"line": 3,
	"column": 12,
	"message": "no mixin named invalid-function\nBacktrace:\n\tstdin:3"
}
*/

// Error reads the original libsass error and creates helpful debuggin
// information for debuggin that error.
func (ctx *Context) ProcessSassError(bs []byte) string {

	if len(bs) == 0 {
		return ""
	}

	err := json.Unmarshal(bs, &ctx.Errors)
	if err != nil {
		log.Fatal(err)
	}

	return string(bs)

	/*
		lines := bytes.Split(ctx.Parser.Output, []byte("\n"))
			// Line number is off by one from libsass
			// Find previous lines to maximum available
			errLines := "" //"error in " + ctx.Parser.LookupFile(pos)
			red := color.NewStyle(color.BlackPaint, color.RedPaint).Brush()
			first := pos - 7
			if first < 0 {
				first = 0
			}
			last := pos + 7
			if last > len(lines) {
				last = len(lines)
			}
			for i := first; i < last; i++ {
				// translate 0 index to 1 index
				str := fmt.Sprintf("\n%3d: %.80s", i+1, lines[i])
				if i == pos-1 {
					str = red(str)
				}
				errLines += str
			}
			ctx.errorString = s + "\n" + errLines
	*/
}

func (ctx *Context) Error() string {
	return ctx.errorString
}
