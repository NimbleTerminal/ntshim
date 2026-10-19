package duckdb

import (
	"errors"
	"fmt"
)

// ErrNonJSPlatform is returned by every entry point when the package is
// imported under a GOOS other than js. It exists so that consumers can use
// build-tag patterns like:
//
//	//go:build !js
//	package data
//	import _ "github.com/duckdb/duckdb-go/v2"
//
//	//go:build js
//	package data
//	import _ "nimbleterminal.dev/ntshim/duckdb"
//
// and still run host-side `go test ./...` without panicking on missing
// JS globals.
var ErrNonJSPlatform = errors.New("ntshim/duckdb: only usable under GOOS=js")

// Error is the typed error returned for failures originating in the JS
// bridge (DuckDB-Wasm errors, connection issues, etc.).
type Error struct {
	// Message is the human-readable error from DuckDB or the bridge.
	Message string
	// Code is set if DuckDB-Wasm exposed a structured code; "" otherwise.
	Code string
}

func (e *Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("duckdb: %s (%s)", e.Message, e.Code)
	}
	return fmt.Sprintf("duckdb: %s", e.Message)
}
