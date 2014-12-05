package context

import (
	"log"
	"path/filepath"
)

func init() {

	RegisterHandler("image-url($a)", ImageUrl)
}

// ImageUrl handles calls to Rel()
func ImageUrl(ctx *Context, csv UnionSassValue) UnionSassValue {
	var path string
	err := Unmarshal(csv, &path)
	// This should create and throw a sass error
	if err != nil {
		log.Fatal(err)
	}
	res, _ := Marshal(filepath.Join(ctx.RelativeImage(), path))
	return res
}
