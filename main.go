package main

import (
	"log"
	"math"
	"os"

	"github.com/rainycape/magick"
)

func main() {
	k1, err := magick.DecodeFile("test/139.jpg")
	if err != nil {
		log.Fatal(err)
	}
	k2, err := magick.DecodeFile("test/140.jpg")
	k3, err := magick.DecodeFile("test/140.jpg")
	k4, err := magick.DecodeFile("test/140.jpg")

	output := composite(k1, k2, k3, k4)

	fo, err := os.Create("output.jpg")
	if err != nil {
		panic(err)
	}
	info := magick.NewInfo()
	info.SetFormat("JPEG")
	err = output.Encode(fo, info)
	if err != nil {
		log.Fatal(err)
	}
	defer fo.Close()
}

func composite(image *magick.Image, rest ...*magick.Image) *magick.Image {
	var images []*magick.Image
	images = append(images, image)
	images = append(images, rest...)

	maxHeight, maxWidth := 0, 0
	//Stack the images vertically determining the combined height
	// and max width necessary
	for i, _ := range images {
		maxHeight += images[i].Height()
		maxWidth = int(math.Max(float64(maxWidth), float64(images[i].Width())))
	}

	output, _ := magick.New(maxWidth, maxHeight)
	currentHeight := 0

	for _, image := range images {
		err := output.Composite(magick.CompositeCopy, image, 0, currentHeight)
		if err != nil {
			panic(err)
		}
		currentHeight += image.Height()
	}

	return output
}
