package sprite_sass

import (
	"fmt"
	"io"
	"log"
)

type Processor struct {
	InFile       io.Reader
	OutFile      io.Writer
	Ipath, Opath string
}

func (p Processor) Run() {
	fmt.Printf("% #v\n", p)
	if p.Ipath == "" || p.Opath == "" {
		log.Fatal("Must provide input and output files")
	}

	parser := Parser{}
	bytes := parser.Start(p.Ipath)
	fmt.Println(string(bytes))
	ctx := SassContext{
		Context: &Context{
			Options: Options{
				OutputStyle:  NESTED_STYLE,
				IncludePaths: make([]string, 0),
			},
			SourceString: string(bytes),
			OutputString: "",
		},
	}
	ctx.Compile()
	fmt.Println("Hello")
	fmt.Println(ctx.Context.OutputString)
}
