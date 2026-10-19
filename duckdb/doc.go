// Package duckdb is a ntshim that exposes DuckDB-Wasm or Haybarn-Wasm
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
// The page selects the engine by loading duckdb-shim.js (DuckDB-Wasm) or
// haybarn-shim.js (Haybarn-Wasm), installed by go tool ntshim-assets --shim=duckdb.
// Await window.ntshim.duckdb.ready before the WASM main runs.
package duckdb
