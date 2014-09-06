package sprite_sass

import (
	"io/ioutil"
	"log"
)

type Processor struct {
	Ipath, Opath string
}

func (p Processor) Run() {

	if p.Ipath == "" || p.Opath == "" {
		log.Fatal("Must provide input and output files")
	}

	parser := Parser{}
	bytes := parser.Start(p.Ipath)

	ctx := Context{
		OutputStyle:  NESTED_STYLE,
		IncludePaths: make([]string, 0),
		Src:          string(bytes),
		Out:          "",
	}

	ctx.Compile()

	err := ioutil.WriteFile(p.Opath, []byte(ctx.Out), 0777)
	if err != nil {
		panic(err)
	}
}
