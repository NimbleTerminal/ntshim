package pdfium

import (
	"errors"
	"fmt"
)

// ErrNonJSPlatform is returned by every entry point when the package is
// imported under a GOOS other than js. It exists so that consumers can
// run host-side `go test ./...` without panicking on missing JS globals.
var ErrNonJSPlatform = errors.New("ntshim/pdfium: only usable under GOOS=js")

// ErrBridgeMissing is returned by Load when the host page did not load
// pdfium-shim.js before the WASM main ran. Unlike ntshim/duckdb
// (which panics on a missing shim), this is an error so consumers can
// degrade gracefully — e.g. a PDF viewer falling back to text mode.
var ErrBridgeMissing = errors.New("ntshim/pdfium: window.ntshim.pdfium is undefined; load pdfium-shim.js before app.wasm")

// Error is the typed error returned for failures originating in the JS
// bridge (PDFium errors, render failures, etc.).
type Error struct {
	// Message is the human-readable error from PDFium or the bridge.
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("pdfium: %s", e.Message)
}
