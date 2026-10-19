# ntshim/duckdb JS bridge

This directory ships in the asset tool. End users do not edit these files
directly — they run `go tool ntshim-assets <web-dir> --shim=duckdb`,
which writes them to `<web-dir>/ntshim/duckdb/`.

Import `duckdb-shim.js` for upstream DuckDB-Wasm or `haybarn-shim.js` for
Haybarn-Wasm. Each entry point uses `duckdb-bridge.js` to install the same
`window.ntshim.duckdb` API and `ready` promise. Load one entry point, and await
`ready` before instantiating `app.wasm`. Haybarn additionally exposes
`prefetchExtensions(names)` for setup-time extension downloads.

Pinned dependencies (bump in the corresponding engine entry point):
- `@duckdb/duckdb-wasm@1.32.0`
- `@haybarn/haybarn-wasm@1.5.6-rc6` (release candidate)
- `apache-arrow@17.0.0`

Haybarn's small browser module is vendored under `haybarn/`, with its bare
Arrow imports patched to the pinned browser CDN module. Its license, source
commit and upstream/patched checksums are recorded alongside it. The entry
point uses explicit worker and WASM URLs on unpkg. Do not use its `getJsDelivrBundles()`:
the package exceeds jsDelivr's size limit. `--vendored` omits the CDN-only
Haybarn entry point and retains the existing DuckDB vendored stub.

To update Haybarn, fetch the pinned npm package's `dist/duckdb-browser.mjs`,
record its upstream checksum, replace both bare `apache-arrow` imports with
`https://cdn.jsdelivr.net/npm/apache-arrow@17.0.0/+esm`, then update the version
in `haybarn-shim.js` and the source/checksums in `haybarn/PROVENANCE.txt`.
Copy the module, provenance and license to the embedded mirror as well.

Keep the runtime, both entry points and the TypeScript definitions in sync
with `cmd/ntshim-assets/assets/duckdb/`. Run the bridge contract tests with
`node --test web/duckdb/bridge.test.mjs`; real Go/WASM examples are documented
in [duckdb/README.md](../../duckdb/README.md).
