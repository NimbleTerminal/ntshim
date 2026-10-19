// Package ntshim is the root of the nimbleterminal.dev/ntshim
// module. The module hosts browser shims that Go-WASM programs hosted by
// github.com/NimbleMarkets/go-booba commonly need.
//
// Each shim is a subpackage with its own Go side (build-tagged for js) and
// an optional JS bridge file under web/<shim>/. The cmd/ntshim-assets tool
// copies the JS side into a consumer's site directory. Audio uses Web Audio
// directly from Go and requires no JS bridge.
//
// First shim: ntshim/duckdb (DuckDB-Wasm via database/sql + Arrow).
package ntshim
