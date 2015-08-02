package types

import "github.com/wellington/spritewell"

type Payloader interface {
	spritewell.Imager
	spritewell.Spriter
}
