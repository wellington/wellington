package wellington

import (
	libsass "github.com/wellington/go-libsass"
	"github.com/wellington/spritewell"
)

// InitializeContext sets up some data structures necessary
// to use wellington
func InitializeContext(ctx *libsass.Context) {
	ctx.Payload = newPayload()
}

type payload struct {
	s spritewell.SafeImageMap
	i spritewell.SafeImageMap
}

func newPayload() payload {
	return payload{
		s: spritewell.NewImageMap(),
		i: spritewell.NewImageMap(),
	}
}

func (p payload) Sprite() spritewell.SafeImageMap {
	return p.s
}

func (p payload) Image() spritewell.SafeImageMap {
	return p.i
}
