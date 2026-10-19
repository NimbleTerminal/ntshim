// Installs window.ntshim.duckdb using upstream DuckDB-Wasm (the default).
// See haybarn-shim.js for the alternative Haybarn engine.
import * as duckdb from 'https://cdn.jsdelivr.net/npm/@duckdb/duckdb-wasm@1.32.0/+esm';
import { tableToIPC } from 'https://cdn.jsdelivr.net/npm/apache-arrow@17.0.0/+esm';
import { installDuckDBBridge } from './duckdb-bridge.js';

installDuckDBBridge(duckdb, tableToIPC, {
    engine: 'duckdb',
    bundles: duckdb.getJsDelivrBundles(),
});
