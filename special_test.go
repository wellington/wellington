package sprite_sass_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"testing"

	. "github.com/drewwells/sprite_sass"
)

func TestInclude(t *testing.T) {

	iString := bytes.NewBufferString(`
$s = sprite-map('*.png')
div {
    background: sprite($s, 139);
    @include sprite-dimensions($s, 139);
    @include background(inline-image('139.png') top left repeat-x);
}

`)
	ctx := Context{}
	r, w := io.Pipe()
	go func(ctx Context, w io.WriteCloser) {
		ctx.Run(iString, w, "test")
		w.Close()
	}(ctx, w)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
