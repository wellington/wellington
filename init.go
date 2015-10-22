package wellington

import "github.com/wellington/spritewell"

type payload struct {
	s *spritewell.SafeImageMap
	i *spritewell.SafeImageMap
}

func newPayload() payload {
	return payload{
		s: spritewell.NewImageMap(),
		i: spritewell.NewImageMap(),
	}
}

func (p payload) Sprite() *spritewell.SafeImageMap {
	return p.s
}

func (p payload) Image() *spritewell.SafeImageMap {
	return p.i
}
