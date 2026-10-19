// Shared Go bridge for API-compatible DuckDB-Wasm distributions.
// Each page imports one engine entry point; only that engine is downloaded.
export function installDuckDBBridge(duckdb, tableToIPC, { engine, bundles }) {
    const namespace = (globalThis.ntshim = globalThis.ntshim || {});
    if (namespace.duckdb) {
        throw new Error('ntshim/duckdb: an engine is already installed; load only one engine entry point');
    }
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
        const bundle = await duckdb.selectBundle(bundles);
        const workerURL = URL.createObjectURL(new Blob(
            [`importScripts("${bundle.mainWorker}");`],
            { type: 'text/javascript' },
        ));
        const worker = new Worker(workerURL);
        const logger = new duckdb.ConsoleLogger(duckdb.LogLevel.WARNING);
        const instance = new duckdb.AsyncDuckDB(logger, worker);
        try {
            await instance.instantiate(bundle.mainModule, bundle.pthreadWorker);
            await instance.open({});
            db = instance;
        } catch (err) {
            worker.terminate();
            throw err;
        } finally {
            URL.revokeObjectURL(workerURL);
        }
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
                // ATTACH/USE failed; release the connection before rethrowing.
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

    const shim = {
        engine,
        open, close, query, exec, registerFileURL, registerFileBuffer,
        ready: bootstrap(),
    };
    if (engine === 'haybarn') {
        let prefetching = false;
        shim.prefetchExtensions = async (names) => {
            await shim.ready;
            if (prefetching) throw new Error('haybarn: extension prefetch is already running');
            prefetching = true;
            try {
                await db.prefetchExtensions(names);
            } finally {
                prefetching = false;
            }
        };
    }
    namespace.duckdb = shim;
    return shim;
}
