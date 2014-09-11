package sprite_sass

import (
	"io"
	"log"
	"os"
)

var (
	Debug *log.Logger
)

func init() {
	// Currently hardwiring log to stderr.  This should
	// be configurable by cli flags.
	Init(os.Stderr)
}

func Init(handle io.Writer) {
	Debug = log.New(handle, "", log.Lshortfile)
}
