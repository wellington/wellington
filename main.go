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

func (l ImageList) X(pos int) int {
	x := 0
	if !l.Vertical {
		return 0
	}
	for i := 0; i < pos; i++ {
		x += l.Is[i].Width()
	}
	return x
}

func (l ImageList) Y(pos int) int {
	y := 0
	for i := 0; i < pos; i++ {
		y += l.Is[i].Height()
	}
	return y
}

func (l *ImageList) Height(sum bool) int {
	h := 0
	ll := *l

	for _, img := range ll.Is {
		if sum && l.Vertical {
			h += img.Height()
		} else {
			h = int(math.Max(float64(h), float64(img.Height())))
		}
	}
	return h
}

func (l *ImageList) Width(sum bool) int {
	w := 0
	ll := *l

	for _, img := range ll.Is {
		if sum && !l.Vertical {
			w += img.Width()
		} else {
			w = int(math.Max(float64(w), float64(img.Width())))
		}
	}
	return w
}

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

func (l *ImageList) Combine(vertical bool) *magick.Image {
	l.Vertical = vertical
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

	curH, curW := 0, 0
	ll := *l
	for _, img := range ll.Is {
		err := out.Composite(magick.CompositeCopy, img, curW, curH)
		if err != nil {
			panic(err)
		}
		if vertical {
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
