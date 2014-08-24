package main

import (
	"os"
	"testing"

	"github.com/rainycape/magick"
)

func TestCombine(t *testing.T) {
	imgs := ImageList{}
	imgs.Decode("test/139.jpg", "test/140.jpg")

	if imgs.Height(true) != 279 {
		t.Errorf("Invalid Height found %d, wanted %d", imgs.Height(true), 279)
	}

	if imgs.Width(true) != 192 {
		t.Errorf("Invalid Height found %d, wanted %d", imgs.Width(true), 192)
	}

	out := imgs.Combine(true)
	fo, err := os.Create("test/output.jpg")
	defer fo.Close()
	if err != nil {
		panic(err)
	}
	jpg := magick.NewInfo()
	jpg.SetFormat("JPEG")
	err = out.Encode(fo, jpg)
	if err != nil {
		panic(err)
	}
}
