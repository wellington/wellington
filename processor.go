package sprite_sass

import (
	"io"
	"io/ioutil"
	"log"
)

type Processor struct {
	InFile       io.Reader
	OutFile      io.Writer
	Ipath, Opath string
}

func (p Processor) Run() {

	if p.Ipath == "" || p.Opath == "" {
		log.Fatal("Must provide input and output files")
	}

	parser := Parser{}
	bytes := parser.Start(p.Ipath)

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

	err := ioutil.WriteFile(p.Opath, []byte(ctx.Context.OutputString), 0777)
	if err != nil {
		panic(err)
	}
}
