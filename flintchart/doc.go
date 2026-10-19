// Package flintchart is a ntshim that exposes the vendored
// flint-ntcharts browser bundle (a compiler from flint ChartAssemblyInput
// JSON to an ntcharts render envelope) to Go-WASM programs.
//
// Under GOOS=js:
//
//	out, err := flintchart.Compile(ctx, inputJSON)
//	// out is the raw envelope JSON: {"spec",...} on success,
//	// {"error":{"message"}} on a compile failure (returned as data,
//	// not error — flint-ntcharts/envelope.Parse surfaces the latter
//	// as a Go error).
//
// Unlike pdfium and duckdb, the bridge is a synchronous, stateless
// function: there is no wasm module to instantiate and no handle to
// manage. Strings cross the syscall/js boundary as JS strings — the
// envelope is JSON text, so no Uint8Array marshaling is needed.
//
// On any other GOOS, every public entry point returns ErrNonJSPlatform
// so host-side test binaries can import the package without panicking.
//
// The page must load ntshim/flintchart/flintchart-shim.js (via
// go tool ntshim-assets) before the WASM main runs; otherwise
// Compile and Version return ErrBridgeMissing.
package flintchart
