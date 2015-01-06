package wellington

import (
	"io"
	"os"
)

func fileReader(path string) io.Reader {
	reader, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	return reader
}
