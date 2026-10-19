//go:build js

package flintchart

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

// ensureReady resolves window.ntshim.flintchart and awaits its ready
// promise. Idempotent: the bundle only needs to be resolved once per
// process. Returns ErrBridgeMissing when the page did not load
// flintchart-shim.js before the WASM main ran.
func ensureReady() (js.Value, error) {
	bridgeOnce.Do(func() {
		root := js.Global().Get("ntshim")
		if root.IsUndefined() {
			bridgeErr = ErrBridgeMissing
			return
		}
		shim := root.Get("flintchart")
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
// *Error. The bridge prefixes its own messages with "flintchart: ", which
// Error() re-adds — trim it to avoid "flintchart: flintchart: ...".
func mapJSError(v js.Value) error {
	if v.IsUndefined() || v.IsNull() {
		return &Error{Message: "unknown rejection"}
	}
	msg := v.Get("message")
	if msg.IsUndefined() {
		return &Error{Message: strings.TrimPrefix(v.String(), "flintchart: ")}
	}
	return &Error{Message: strings.TrimPrefix(msg.String(), "flintchart: ")}
}

// callSync invokes a synchronous bridge function, converting a JS throw
// (which syscall/js surfaces as a Go panic) into a *Error. The compile
// function never throws in practice (compile errors are caught bundle-
// side and returned as an {"error":...} envelope), but a broken bundle
// or an OOM could still throw synchronously, so guard it the same way
// pdfium guards its synchronous calls.
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
