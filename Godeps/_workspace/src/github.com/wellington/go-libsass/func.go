package context

import (
	"unsafe"

	"github.com/wellington/go-libsass/libs"
)

// HandlerFunc describes the method signature for registering
// a Go function to be called by libsass.
type HandlerFunc func(v interface{}, req SassValue, res *SassValue) error

// Handler accepts a HandlerFunc and returns SassCallback for sending
// to libsass. The third argument must be a pointer and the function
// must return an error.
func Handler(h HandlerFunc) libs.SassCallback {
	return func(v interface{}, usv libs.UnionSassValue, rsv *libs.UnionSassValue) error {
		if *rsv == nil {
			*rsv = libs.MakeNil()
		}
		req := SassValue{value: usv}
		res := SassValue{value: *rsv}
		err := h(v, req, &res)

		// FIXME: This shouldn't be happening, handler should assign
		// to the address properly.
		*rsv = res.Val()

		return err
	}
}

type handler struct {
	sign     string
	callback libs.SassCallback
}

// RegisterHandler sets the passed signature and callback to the
// handlers array.
func RegisterHandler(sign string, callback HandlerFunc) {
	handlers = append(handlers, handler{
		sign:     sign,
		callback: Handler(callback),
	})
}

var _ libs.SassCallback = TestCallback

func testCallback(h HandlerFunc) libs.SassCallback {
	return func(v interface{}, _ libs.UnionSassValue, _ *libs.UnionSassValue) error {
		return nil
	}
}

// TestCallback implements libs.SassCallback. TestCallback is a useful
// place to start when developing new handlers.
var TestCallback = testCallback(func(_ interface{}, _ SassValue, _ *SassValue) error {
	return nil
})

// handlers is the list of registered sass handlers
var handlers []handler

// Cookie is used for passing context information to libsass.  Cookie is
// passed to custom handlers when libsass executes them through the go
// bridge.
type Cookie struct {
	Sign string
	Fn   libs.SassCallback
	Ctx  interface{}
}

// SetFunc assigns the registered methods to SassOptions. Functions
// are called when the compiler encounters the registered signature.
func (ctx *Context) SetFunc(goopts libs.SassOptions) {
	cookies := make([]libs.Cookie, len(handlers)+len(ctx.Cookies))
	// Append registered handlers to cookie array
	for i, h := range handlers {
		cookies[i] = libs.Cookie{
			Sign: h.sign,
			Fn:   h.callback,
			Ctx:  ctx,
		}
	}
	for i, h := range ctx.Cookies {
		cookies[i+len(handlers)] = libs.Cookie{
			Sign: h.Sign,
			Fn:   h.Fn,
			Ctx:  ctx,
		}
	}
	// TODO: this seems to run fine with garbage collection on
	// surprisingly enough
	// disable garbage collection of cookies. These need to
	// be manually freed in the wrapper
	gofns := make([]libs.SassFunc, len(cookies))
	for i, cookie := range cookies {
		fn := libs.SassMakeFunction(cookie.Sign,
			unsafe.Pointer(&cookies[i]))
		gofns[i] = fn
	}
	libs.BindFuncs(goopts, gofns)
}
