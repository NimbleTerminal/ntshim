//go:build js

package pdfium

import (
	"strings"
	"sync"
	"syscall/js"
)

var (
	bridgeOnce sync.Once
	bridgeErr  error
	bridgeVal  js.Value
)

// ensureReady resolves window.ntshim.pdfium and awaits its ready
// promise. Idempotent: the PDFium wasm download/instantiate happens
// once per process. Returns ErrBridgeMissing when the page did not
// load pdfium-shim.js before the WASM main ran.
func ensureReady() (js.Value, error) {
	bridgeOnce.Do(func() {
		root := js.Global().Get("ntshim")
		if root.IsUndefined() {
			bridgeErr = ErrBridgeMissing
			return
		}
		shim := root.Get("pdfium")
		if shim.IsUndefined() {
			bridgeErr = ErrBridgeMissing
			return
		}
		if _, err := awaitPromise(shim.Get("ready")); err != nil {
			bridgeErr = err
			return
		}
		bridgeVal = shim
	})
	return bridgeVal, bridgeErr
}

// awaitPromise blocks the calling goroutine until p resolves or rejects.
// Go's WASM scheduler yields to JS while the goroutine is parked on the
// channel, so JS callbacks fire and the promise settles.
//
// On reject, returns a *Error whose Message is the rejected value's
// .message (or its String() form if .message is missing).
func awaitPromise(p js.Value) (js.Value, error) {
	done := make(chan struct{})
	var resolved, rejected js.Value
	var ok bool

	thenFn := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			resolved = args[0]
		}
		ok = true
		close(done)
		return nil
	})
	catchFn := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			rejected = args[0]
		}
		close(done)
		return nil
	})
	defer thenFn.Release()
	defer catchFn.Release()

	p.Call("then", thenFn).Call("catch", catchFn)
	<-done

	if !ok {
		return js.Value{}, mapJSError(rejected)
	}
	return resolved, nil
}

// mapJSError converts a thrown/rejected JS Error (or any value) into a
// *Error. The bridge prefixes its own messages with "pdfium: ", which
// Error() re-adds — trim it to avoid "pdfium: pdfium: ...".
func mapJSError(v js.Value) error {
	if v.IsUndefined() || v.IsNull() {
		return &Error{Message: "unknown rejection"}
	}
	msg := v.Get("message")
	if msg.IsUndefined() {
		return &Error{Message: strings.TrimPrefix(v.String(), "pdfium: ")}
	}
	return &Error{Message: strings.TrimPrefix(msg.String(), "pdfium: ")}
}

// callSync invokes a synchronous bridge function, converting a JS throw
// (which syscall/js surfaces as a Go panic) into a *Error. duckdb's
// bridge calls are all Promise-based so failures arrive as rejections
// via awaitPromise; pdfium's numPages/pageSize/closeDocument throw
// synchronously instead.
func callSync(recv js.Value, method string, args ...any) (res js.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case js.Error:
				err = mapJSError(e.Value)
			case *js.Error:
				err = mapJSError(e.Value)
			default:
				panic(r)
			}
		}
	}()
	return recv.Call(method, args...), nil
}

// goBytesToJS allocates a Uint8Array on the JS side and copies src into
// it. The returned js.Value is owned by the JS GC.
func goBytesToJS(src []byte) js.Value {
	u8 := js.Global().Get("Uint8Array").New(len(src))
	js.CopyBytesToJS(u8, src)
	return u8
}
