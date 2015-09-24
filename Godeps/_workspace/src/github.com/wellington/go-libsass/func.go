package libsass

import (
	"fmt"
	"sync"

	"github.com/wellington/go-libsass/libs"
)

var ghMu sync.RWMutex

// globalHandlers is the list of predefined handlers registered externally
var globalHandlers []handler

// RegisterHandler sets the passed signature and callback to the
// handlers array.
func RegisterHandler(sign string, callback HandlerFunc) {
	ghMu.Lock()
	globalHandlers = append(globalHandlers, handler{
		sign:     sign,
		callback: Handler(callback),
	})
	ghMu.Unlock()
}

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
		if rsv != nil {
			*rsv = res.Val()
		}

		return err
	}
}

type handler struct {
	sign     string
	callback libs.SassCallback
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

// Cookie is used for passing context information to libsass.  Cookie is
// passed to custom handlers when libsass executes them through the go
// bridge.
type Func struct {
	Sign string
	Fn   libs.SassCallback
	Ctx  interface{}
}

type Funcs struct {
	sync.RWMutex
	wg      sync.WaitGroup
	closing chan struct{}
	f       []Func
	idx     []*string

	// Func are complex, we need a reference to the entire context to
	// successfully execute them.
	ctx *Context
}

func NewFuncs(ctx *Context) *Funcs {
	return &Funcs{
		closing: make(chan struct{}),
		ctx:     ctx,
	}
}

func (fs *Funcs) Add(f Func) {
	fs.Lock()
	defer fs.Unlock()
	fs.f = append(fs.f, f)
}

// SetFunc assigns the registered methods to SassOptions. Functions
// are called when the compiler encounters the registered signature.
func (fs *Funcs) Bind(goopts libs.SassOptions) {
	ghMu.RLock()
	cookies := make([]libs.Cookie, len(globalHandlers)+len(fs.f))
	// Append registered handlers to cookie array
	for i, h := range globalHandlers {
		cookies[i] = libs.Cookie{
			Sign: h.sign,
			Fn:   h.callback,
			Ctx:  fs.ctx,
		}
	}
	l := len(globalHandlers)
	ghMu.RUnlock()

	for i, h := range fs.f {
		cookies[i+l] = libs.Cookie{
			Sign: h.Sign,
			Fn:   h.Fn,
			Ctx:  fs.ctx,
		}
	}
	fs.idx = libs.BindFuncs(goopts, cookies)
}

func (fs *Funcs) Close() {
	err := libs.RemoveFuncs(fs.idx)
	if err != nil {
		fmt.Println("error cleaning up funcs", err)
	}
	close(fs.closing)
	fs.wg.Wait()
}
