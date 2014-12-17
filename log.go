package wellington

import (
	"io"
	"log"
	"os"
)

var (
	// Debug future ability to toggle the logging level and destination.
	Debug *log.Logger
)

func init() {
	// Currently hardwiring log to stderr.  This should
	// be configurable by cli flags.
	Init(os.Stderr)
}

// Init setups an application logger
func Init(handle io.Writer) {
	Debug = log.New(handle, "", log.Lshortfile)
}
