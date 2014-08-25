package main

import (
	"log"

	"github.com/moovweb/gosass"
)

type Context struct {
	Sass          *gosass.FileContext
	Sprites       []ImageList
	Input, Output string
}

func (ctx Context) Compile() {
	if ctx.Sass.InputPath == "" || ctx.Output == "" {
		log.Fatal("No input/output file specified")
	}
	gosass.CompileFile(ctx.Sass)

}

func (ctx Context) Export() {
}
