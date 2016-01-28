// Create a preamble for every CSS file
package main

import (
	"bytes"
	"log"
	"os"

	"github.com/wellington/go-libsass"
)

func main() {
	libsass.RegisterHeader(`
/*
 * This is a preamble which will be added to every CSS file generated.
 * It uses the libSass header which includes the text on every file.
 *
 */`)

	// Input Sass source
	input := `div {}`

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
	// /*
	//  * This is a preamble which will be added to every CSS file generated.
	//	* It uses the libSass header which includes the text on every file.
	//	*
	//	*/
	//
}
