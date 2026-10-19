// ntshim/duckdb — JS bridge file
//
// Loaded by consumer pages before their Go-WASM main runs. Sets
// window.ntshim.duckdb with the API the Go side expects. Reads Arrow
// IPC out of DuckDB-Wasm's native arrow.Table results and ships bytes
// across the JS↔Go boundary.

import * as duckdb from 'https://cdn.jsdelivr.net/npm/@duckdb/duckdb-wasm@1.32.0/+esm';
import { tableToIPC } from 'https://cdn.jsdelivr.net/npm/apache-arrow@17.0.0/+esm';

const NAMESPACE = (window.ntshim = window.ntshim || {});

const connections = new Map();
let nextHandle = 1;
let db = null;

function parseDSN(dsn) {
    const [path, query = ''] = dsn.split('?');
    const opts = {};
    for (const part of query.split('&').filter(Boolean)) {
        const [k, v = ''] = part.split('=');
        opts[decodeURIComponent(k)] = decodeURIComponent(v);
    }
    return { path, opts };
}

async function bootstrap() {
    const bundles = duckdb.getJsDelivrBundles();
    const bundle = await duckdb.selectBundle(bundles);
    const workerURL = URL.createObjectURL(new Blob(
        [`importScripts("${bundle.mainWorker}");`],
        { type: 'text/javascript' },
    ));
    const worker = new Worker(workerURL);
    const logger = new duckdb.ConsoleLogger(duckdb.LogLevel.WARNING);
    db = new duckdb.AsyncDuckDB(logger, worker);
    await db.instantiate(bundle.mainModule, bundle.pthreadWorker);
    URL.revokeObjectURL(workerURL);
}

async function open(dsn) {
    if (!db) throw new Error('duckdb: bridge not bootstrapped (await ntshim.duckdb.ready first)');
    const { path, opts } = parseDSN(dsn);
    const conn = await db.connect();
    if (path && path !== ':memory:') {
        const readOnly = opts.access_mode === 'read_only';
        const sql = `ATTACH '${path.replace(/'/g, "''")}' AS attached${readOnly ? ' (READ_ONLY)' : ''}`;
        try {
            await conn.query(sql);
            await conn.query('USE attached');
        } catch (err) {
            // ATTACH/USE failed — close the orphaned connection before
            // rethrowing so it isn't leaked in DuckDB-Wasm (it was never
            // stored in `connections`, so closeDocument can't reach it).
            await conn.close();
            throw err;
        }
    }
    const handle = nextHandle++;
    connections.set(handle, conn);
    return handle;
}

async function close(handle) {
    const conn = connections.get(handle);
    if (!conn) return;
    await conn.close();
    connections.delete(handle);
}

function getConn(handle) {
    const conn = connections.get(handle);
    if (!conn) throw new Error(`duckdb: invalid connection handle ${handle}`);
    return conn;
}

async function query(handle, sql, params) {
    const conn = getConn(handle);
    const stmt = await conn.prepare(sql);
    try {
        const table = (params && params.length > 0)
            ? await stmt.query(...params)
            : await stmt.query();
        return tableToIPC(table, 'stream');
    } finally {
        await stmt.close();
    }
}

async function exec(handle, sql, params) {
    const conn = getConn(handle);
    const stmt = await conn.prepare(sql);
    try {
        if (params && params.length > 0) {
            await stmt.query(...params);
        } else {
            await stmt.query();
        }
        // DuckDB-Wasm does not surface rowsAffected; report 0.
        return { rowsAffected: 0 };
    } finally {
        await stmt.close();
    }
}

async function registerFileURL(name, url) {
    if (!db) throw new Error('duckdb: bridge not bootstrapped');
    await db.registerFileURL(name, url, duckdb.DuckDBDataProtocol.HTTP, false);
}

async function registerFileBuffer(name, bytes) {
    if (!db) throw new Error('duckdb: bridge not bootstrapped');
    await db.registerFileBuffer(name, bytes);
}

NAMESPACE.duckdb = {
    open, close, query, exec, registerFileURL, registerFileBuffer,
    ready: bootstrap(),
};
