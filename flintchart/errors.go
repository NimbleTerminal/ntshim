package flintchart

import (
	"errors"
	"fmt"
)

// ErrNonJSPlatform is returned by every entry point when the package is
// imported under a GOOS other than js. It exists so that consumers can
// run host-side `go test ./...` without panicking on missing JS globals.
var ErrNonJSPlatform = errors.New("ntshim/flintchart: only usable under GOOS=js")

// ErrBridgeMissing is returned when the host page did not load
// flintchart-shim.js before the WASM main ran.
var ErrBridgeMissing = errors.New("ntshim/flintchart: window.ntshim.flintchart is undefined; load flintchart-shim.js before app.wasm")

// Error is the typed error returned for failures originating in the JS
// bridge (a thrown/rejected value — a broken bundle, OOM, etc.). Compile
// failures reported by the compiler itself are NOT surfaced this way:
// they come back as a successful []byte envelope (`{"error":{"message"}}`)
// for the caller to parse as data.
type Error struct {
	// Message is the human-readable error from the bridge.
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("flintchart: %s", e.Message)
}
