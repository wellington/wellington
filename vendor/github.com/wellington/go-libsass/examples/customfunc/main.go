package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"
	"os"

	"golang.org/x/net/context"

	"github.com/wellington/go-libsass"
)

// cryptotext is of type libsass.SassFunc. As libsass compiles the
// source Sass, it will look for `crypto()` then call this function.
func cryptotext(ctx context.Context, usv libsass.SassValue) (*libsass.SassValue, error) {

	c := 10
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	res, err := libsass.Marshal(fmt.Sprintf("'%x'", b))
	return &res, err
}

func main() {
	// Register a custom Sass func crypto
	libsass.RegisterSassFunc("crypto()", cryptotext)

	// Input Sass source
	input := `div { text: crypto(); }`

	buf := bytes.NewBufferString(input)
	// Starts a compiler writing to Stdout and reading from
	// a bytes buffer of the input source
	comp, err := libsass.New(os.Stdout, buf)
	if err != nil {
		log.Fatal(err)
	}

	// Run() kicks off the compiler and instructs libsass how
	// to read our input, handling the output from libsass
	if err := comp.Run(); err != nil {
		log.Fatal(err)
	}

	// Output:
	// div {
	//   text: 'c91db27d5e580ef4292e'; }
}
