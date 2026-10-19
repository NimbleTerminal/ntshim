// ntshim/duckdb — JS bridge (vendored variant) — STUB
//
// VENDORED MODE IS NOT SUPPORTED IN v0.1.0.
//
// Why: The DuckDB-Wasm npm dist (duckdb-browser.mjs) uses a bare module
// specifier `import * as u from "apache-arrow"` internally, which cannot
// be resolved from a local file path without an ES module import map.
// Additionally, getJsDelivrBundles() inside duckdb-browser.mjs hard-codes
// https://cdn.jsdelivr.net/... URLs for the .wasm and worker files.
// The JSDelivr +esm transform likewise references /npm/apache-arrow@17.0.0/+esm
// as an absolute CDN path. There is no single-file fully-self-contained ESM
// bundle available for duckdb-wasm@1.32.0 that works from local file paths
// without an import map.
//
// A future task can resolve this by either:
//   a) Adding an importmap <script> to the consumer HTML that maps
//      "apache-arrow" → "./vendored/apache-arrow-bundled.mjs" and uses
//      a patched duckdb-browser.mjs that reads wasm/worker from local paths.
//   b) Producing a single fully-inlined bundle with a custom esbuild step.
//
// The vendored bundles (duckdb-mvp.wasm, duckdb-eh.wasm, worker .js files,
// duckdb-browser.mjs, apache-arrow.mjs) ARE present in the vendored/
// subdirectory and embedded in the asset tool — they just cannot be wired
// up without the above infrastructure.
//
// USE --cdn (the default) for production. It is the verified working path.

(window.ntshim = window.ntshim || {});
window.ntshim.duckdb = {
    open: async () => {
        throw new Error(
            'ntshim/duckdb: vendored mode is a v0.1.0 stub. ' +
            'Use --cdn (the default) for production. ' +
            'See web/duckdb/duckdb-shim-vendored.js for details.'
        );
    },
    close: async () => { throw new Error('ntshim/duckdb: vendored mode is a v0.1.0 stub. Use --cdn.'); },
    query: async () => {
        throw new Error('ntshim/duckdb: vendored mode stub — use --cdn');
    },
    exec: async () => {
        throw new Error('ntshim/duckdb: vendored mode stub — use --cdn');
    },
    registerFileURL: async () => {
        throw new Error('ntshim/duckdb: vendored mode stub — use --cdn');
    },
    registerFileBuffer: async () => {
        throw new Error('ntshim/duckdb: vendored mode stub — use --cdn');
    },
    ready: Promise.resolve(),
};
