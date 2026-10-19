//go:build js

package duckdb

import (
	"syscall/js"
)

// bridge returns the window.ntshim.duckdb object. It panics if the
// page did not load an engine entry point before the WASM main ran — that's
// strictly a setup error and there is no useful recovery.
func bridge() js.Value {
	root := js.Global().Get("ntshim")
	if root.IsUndefined() {
		panic("ntshim/duckdb: window.ntshim is undefined; load duckdb-shim.js or haybarn-shim.js before app.wasm")
	}
	shim := root.Get("duckdb")
	if shim.IsUndefined() {
		panic("ntshim/duckdb: window.ntshim.duckdb is undefined; load duckdb-shim.js or haybarn-shim.js before app.wasm")
	}
	return shim
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

// mapJSError converts a rejected JS Error (or any value) into a *Error.
func mapJSError(v js.Value) error {
	if v.IsUndefined() || v.IsNull() {
		return &Error{Message: "unknown rejection"}
	}
	msg := v.Get("message")
	if msg.IsUndefined() {
		return &Error{Message: v.String()}
	}
	return &Error{Message: msg.String()}
}
