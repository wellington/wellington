package main

import (
	"math"

	"github.com/rainycape/magick"
)

type Is []*magick.Image

type ImageList struct {
	Is
	Vertical bool
}

// Return the X position of an image based
// on the layout (vertical/horizontal) and
// position in Image slice
func (l ImageList) X(pos int) int {
	x := 0
	if l.Vertical {
		return 0
	}
	for i := 0; i < pos; i++ {
		x += l.Is[i].Width()
	}
	return x
}

// Return the Y position of an image based
// on the layout (vertical/horizontal) and
// position in Image slice
func (l ImageList) Y(pos int) int {
	y := 0
	if !l.Vertical {
		return 0
	}
	for i := 0; i < pos; i++ {
		y += l.Is[i].Height()
	}
	return y
}

// Return the cumulative Height of the
// image slice.
func (l *ImageList) Height() int {
	h := 0
	ll := *l

	for _, img := range ll.Is {
		if l.Vertical {
			h += img.Height()
		} else {
			h = int(math.Max(float64(h), float64(img.Height())))
		}
	}
	return h
}

// Return the cumulative Width of the
// image slice.
func (l *ImageList) Width() int {
	w := 0
	ll := *l

	for _, img := range ll.Is {
		if !l.Vertical {
			w += img.Width()
		} else {
			w = int(math.Max(float64(w), float64(img.Width())))
		}
	}
	return w
}

// Accept a variable number of image paths returning
// an image slice of each file path decoded into a
// *magick.Image.
func (l *ImageList) Decode(rest ...string) {
	ll := *l
	for _, path := range rest {
		img, err := magick.DecodeFile(path)
		if err != nil {
			panic(err)
		}
		ll.Is = append(ll.Is, img)
	}
	*l = ll
}

// Combine all images in the slice into a final output
// image.
func (l *ImageList) Combine() *magick.Image {

	var (
		out        *magick.Image
		maxW, maxH int
	)

	maxW, maxH = l.Width(), l.Height()

	out, _ = magick.New(maxW, maxH)

	curH, curW := 0, 0
	ll := *l
	for _, img := range ll.Is {
		err := out.Composite(magick.CompositeCopy, img, curW, curH)
		if err != nil {
			panic(err)
		}
		if ll.Vertical {
			curH += img.Height()
		} else {
			curW += img.Width()
		}
	}

	l = &ll

	return out
}

func main() {
}
