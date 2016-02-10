package payload

import (
	"github.com/wellington/spritewell"
	"golang.org/x/net/context"
)

// Handles the conversion of payload to context and vice versa.

type key int

const (
	_             = iota
	spriteKey key = iota
	imageKey  key = iota
)

// New returns a Context with an attached payload for Sprites and Images
func New() context.Context {
	ctx := context.WithValue(context.TODO(),
		spriteKey, spritewell.NewImageMap())
	ctx = context.WithValue(ctx,
		imageKey, spritewell.NewImageMap())

	return ctx
}

// Payloader describes the way to communicate with underlying datastore
// a payload describes.
type Payloader interface {
	Get(key string) *spritewell.Sprite
	Set(key string, sprite *spritewell.Sprite)
	ForEach(func(key string, sprite *spritewell.Sprite))
}

// Sprite is a convenience to return Sprite payload
func Sprite(ctx context.Context) Payloader {
	return ctx.Value(spriteKey).(Payloader)
}

// Image is a convenience to return Image payload
func Image(ctx context.Context) Payloader {
	return ctx.Value(imageKey).(Payloader)
}
