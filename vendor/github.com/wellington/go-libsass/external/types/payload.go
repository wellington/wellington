package types

import "github.com/wellington/spritewell"

// Payloader allows a way to add spritewell information to
// the libsass compiler/context. Payload is passed to handlers
// for processing sprite information.
type Payloader interface {
	spritewell.Imager
	spritewell.Spriter
}

type payload struct {
	s *spritewell.SafeImageMap
	i *spritewell.SafeImageMap
}

func NewPayload() Payloader {
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
