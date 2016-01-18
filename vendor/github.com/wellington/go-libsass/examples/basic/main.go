package main

import (
	"log"
	"os"

	libsass "github.com/wellington/go-libsass"
)

func main() {
	r, err := os.Open("file.scss")
	if err != nil {
		log.Fatal(err)
	}
	comp, err := libsass.New(os.Stdout, r)
	if err != nil {
		log.Fatal(err)
	}

	if err := comp.Run(); err != nil {
		log.Fatal(err)
	}
}
