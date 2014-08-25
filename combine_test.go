package main

import (
	"os"
	"testing"

	"github.com/rainycape/magick"
)

func TestCombine(t *testing.T) {
	imgs := ImageList{}
	imgs.Decode("test/139.jpg", "test/140.jpg")
	imgs.Vertical = true
	out := imgs.Combine()
	if imgs.Height() != 279 {
		t.Errorf("Invalid Height found %d, wanted %d", imgs.Height(), 279)
	}

	if imgs.Width() != 96 {
		t.Errorf("Invalid Width found %d, wanted %d", imgs.Width(), 192)
	}

	if x := imgs.X(1); x != 0 {
		t.Errorf("Invalid X found %d, wanted %d", x, 96)
	}

	if y := imgs.Y(1); y != 139 {
		t.Errorf("Invalid Y found %d, wanted %d", y, 139)
	}
	testFile := "test/output.jpg"
	fo, err := os.Create(testFile)
	defer func() {
		fo.Close()
		err := os.Remove(testFile)

		if err != nil {
			panic(err)
		}
	}()
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
