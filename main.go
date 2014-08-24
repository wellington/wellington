package main

import (
	"math"

	"github.com/rainycape/magick"
)

type Image struct {
	*magick.Image
	x, y int
}

type ImageList []*magick.Image

func (l ImageList) Height(sum bool) int {
	h := 0
	for _, img := range l {
		if sum {
			h += img.Height()
		} else {
			h = int(math.Max(float64(h), float64(img.Height())))
		}
	}
	return h
}

func (l ImageList) Width(sum bool) int {
	w := 0
	for _, img := range l {
		if sum {
			w += img.Width()
		} else {
			w = int(math.Max(float64(w), float64(img.Width())))
		}
	}
	return w
}

func (l *ImageList) Decode(rest ...string) {
	for _, path := range rest {
		img, err := magick.DecodeFile(path)
		if err != nil {
			panic(err)
		}
		*l = append(*l, img)
	}
}

func (l *ImageList) Combine(vertical bool) *magick.Image {

	var (
		out        *magick.Image
		maxW, maxH int
	)

	if vertical {
		maxW, maxH = l.Width(false), l.Height(true)
	} else {
		maxW, maxH = l.Width(true), l.Height(false)
	}

	out, _ = magick.New(maxW, maxH)

	currentHeight := 0

	for _, img := range *l {
		err := out.Composite(magick.CompositeCopy, img, 0, currentHeight)
		if err != nil {
			panic(err)
		}
		currentHeight += img.Height()
	}

	return out
}

func main() {
}
