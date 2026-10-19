# ntshim/duckdb JS bridge

This directory ships in the asset tool. End users do not edit these files
directly — they run `go tool ntshim-assets <web-dir> --shim=duckdb`,
which writes them to `<web-dir>/ntshim/duckdb/`.

The bridge sets `window.ntshim.duckdb` and a `ready` promise. The page
must await that promise before instantiating `app.wasm`.

Pinned dependencies (bump in `duckdb-shim.js`):
- `@duckdb/duckdb-wasm@1.32.0`
- `apache-arrow@17.0.0`
