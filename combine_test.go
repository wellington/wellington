package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/rainycape/magick"
)

func TestCombine(t *testing.T) {
	imgs := ImageList{}
	imgs.Decode("test/139.jpg", "test/140.jpg")

	out := imgs.Combine(true)
	if imgs.Height(true) != 279 {
		t.Errorf("Invalid Height found %d, wanted %d", imgs.Height(true), 279)
	}

	if imgs.Height(false) != 140 {
		t.Errorf("Invalid Height found %d, wanted %d", imgs.Height(false), 140)
	}

	if imgs.Width(true) != 96 {
		t.Errorf("Invalid Width found %d, wanted %d", imgs.Width(true), 192)
	}

	if imgs.Width(false) != 96 {
		t.Errorf("Invalid Width found %d, wanted %d", imgs.Width(false), 96)
	}
	return
	if x := imgs.X(1); x != 96 {
		t.Errorf("Invalid Height found %d, wanted %d", x, 96)
	}

	if y := imgs.Y(1); y != 139 {
		t.Errorf("Invalid Height found %d, wanted %d", y, 139)
	}

	fmt.Println(imgs.X(1))

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
