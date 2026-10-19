// Package duckdb is a ntshim that exposes upstream @duckdb/duckdb-wasm
// to Go-WASM programs as a database/sql driver named "duckdb" and a native
// Arrow API.
//
// Under GOOS=js, importing this package for side effects registers the
// driver:
//
//	import _ "nimbleterminal.dev/ntshim/duckdb"
//	db, err := sql.Open("duckdb", "my-data.duckdb?access_mode=read_only")
//
// On any other GOOS, every public entry point returns ErrNonJSPlatform so
// host-side test binaries can import the package without panicking.
//
// The page must load web/duckdb/duckdb-shim.js (via go tool
// ntshim-assets) before the WASM main runs.
package duckdb
