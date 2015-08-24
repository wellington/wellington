package main

import "github.com/wellington/go-libsass/libs"

func main() {
	sassFile := "../../test/sass/error.scss"
	gofc := libs.SassMakeFileContext(sassFile)
	gocc := libs.SassMakeFileCompiler(gofc)

	libs.SassCompilerParse(gocc)

}
