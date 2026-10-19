// Alternative engine: Haybarn-Wasm (release candidate), powered by DuckDB.
// Installs the same window.ntshim.duckdb API as duckdb-shim.js.
// The small vendored module resolves Arrow imports to the pinned browser build.
// Haybarn exceeds jsDelivr's package size limit, so worker/WASM URLs are explicit.
import * as duckdb from './haybarn/duckdb-browser.mjs';
import { tableToIPC } from 'https://cdn.jsdelivr.net/npm/apache-arrow@17.0.0/+esm';
import { installDuckDBBridge } from './duckdb-bridge.js';

const base = 'https://unpkg.com/@haybarn/haybarn-wasm@1.5.6-rc6/dist/';
installDuckDBBridge(duckdb, tableToIPC, {
    engine: 'haybarn',
    bundles: {
        mvp: {
            mainModule: `${base}duckdb-mvp.wasm`,
            mainWorker: `${base}duckdb-browser-mvp.worker.js`,
        },
        eh: {
            mainModule: `${base}duckdb-eh.wasm`,
            mainWorker: `${base}duckdb-browser-eh.worker.js`,
        },
    },
});
