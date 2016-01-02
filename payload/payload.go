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

func New() context.Context {
	ctx := context.WithValue(context.TODO(),
		spriteKey, spritewell.NewImageMap())
	ctx = context.WithValue(ctx,
		imageKey, spritewell.NewImageMap())

	return ctx
}

type Payloader interface {
	Get(key string) *spritewell.Sprite
	Set(key string, sprite *spritewell.Sprite)
	ForEach(func(key string, sprite *spritewell.Sprite))
}

func Sprite(ctx context.Context) Payloader {
	return ctx.Value(spriteKey).(Payloader)
}

func Image(ctx context.Context) Payloader {
	return ctx.Value(spriteKey).(Payloader)
}
