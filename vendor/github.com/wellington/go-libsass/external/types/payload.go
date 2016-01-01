package types

import "github.com/wellington/spritewell"

// Payloader allows a way to add spritewell information to
// the libsass compiler/context. Payload is passed to handlers
// for processing sprite information.
type Payloader interface {
	spritewell.Imager
	spritewell.Spriter
}
