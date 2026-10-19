import assert from 'node:assert/strict';
import { test } from 'node:test';
import { installDuckDBBridge } from './duckdb-bridge.js';

function harness(t, engine = 'duckdb', failures = {}) {
    const previous = {
        namespace: globalThis.ntshim,
        Worker: globalThis.Worker,
        create: URL.createObjectURL,
        revoke: URL.revokeObjectURL,
    };
    delete globalThis.ntshim;
    const calls = [];
    globalThis.Worker = class {
        constructor(url) { calls.push(['worker', url]); }
        terminate() { calls.push(['terminate']); }
    };
    URL.createObjectURL = () => 'blob:test-worker';
    URL.revokeObjectURL = (url) => calls.push(['revoke', url]);
    t.after(() => {
        if (previous.namespace === undefined) delete globalThis.ntshim;
        else globalThis.ntshim = previous.namespace;
        globalThis.Worker = previous.Worker;
        URL.createObjectURL = previous.create;
        URL.revokeObjectURL = previous.revoke;
    });

    const bytes = new Uint8Array([1, 2, 3]);
    const connection = {
        async close() {
            calls.push(['close']);
            if (failures.close) throw new Error('close failed');
        },
        async query(sql) {
            calls.push(['sql', sql]);
            if (failures.attach) throw new Error('attach failed');
        },
        async prepare(sql) {
            calls.push(['prepare', sql]);
            return {
                async query(...params) {
                    calls.push(['query', ...params]);
                    if (failures.query) throw new Error('query failed');
                    return { result: true };
                },
                async close() { calls.push(['statement-close']); },
            };
        },
    };
    const api = {
        async selectBundle(bundles) {
            assert.equal(bundles.eh.mainModule, `${engine}.wasm`);
            return bundles.eh;
        },
        ConsoleLogger: class {},
        LogLevel: { WARNING: 2 },
        DuckDBDataProtocol: { HTTP: 4 },
        AsyncDuckDB: class {
            async instantiate(module) {
                calls.push(['instantiate', module]);
                if (failures.instantiate) throw new Error('instantiate failed');
            }
            async open(config) { calls.push(['database-open', config]); }
            async connect() { return connection; }
            async registerFileURL(...args) { calls.push(['file-url', ...args]); }
            async registerFileBuffer(...args) { calls.push(['file-buffer', ...args]); }
            async prefetchExtensions(names) { calls.push(['prefetch', names]); }
        },
    };
    const serialize = (table, format) => {
        assert.deepEqual(table, { result: true });
        assert.equal(format, 'stream');
        return bytes;
    };
    const shim = installDuckDBBridge(api, serialize, {
        engine,
        bundles: { eh: { mainModule: `${engine}.wasm`, mainWorker: `${engine}.worker.js` } },
    });
    return { shim, calls, bytes, connection, api, serialize };
}

for (const engine of ['duckdb', 'haybarn']) {
    test(`${engine}: SQL parameters, Arrow IPC and file registration`, async (t) => {
        const { shim, calls, bytes } = harness(t, engine);
        await assert.rejects(shim.open(':memory:'), /not bootstrapped/);
        await shim.ready;
        assert.equal(globalThis.ntshim.duckdb, shim);
        assert.equal(shim.engine, engine);
        assert.deepEqual(calls.slice(1, 4), [
            ['instantiate', `${engine}.wasm`], ['database-open', {}], ['revoke', 'blob:test-worker'],
        ]);

        const handle = await shim.open(':memory:');
        assert.equal(await shim.query(handle, 'SELECT ?', [42]), bytes);
        assert.deepEqual(await shim.exec(handle, 'INSERT INTO t VALUES (?)', ['hello']), { rowsAffected: 0 });
        assert(calls.some((c) => c[0] === 'query' && c[1] === 42));
        assert(calls.some((c) => c[0] === 'query' && c[1] === 'hello'));
        assert.equal(calls.filter((c) => c[0] === 'statement-close').length, 2);

        await shim.registerFileURL('data.csv', 'https://example.test/data.csv');
        await shim.registerFileBuffer('data.bin', bytes);
        assert.deepEqual(calls.at(-2), ['file-url', 'data.csv', 'https://example.test/data.csv', 4, false]);
        assert.deepEqual(calls.at(-1), ['file-buffer', 'data.bin', bytes]);

        await shim.close(handle);
        await shim.close(handle); // Closing an already-released handle is harmless.
        assert.equal(calls.filter((c) => c[0] === 'close').length, 1);
        await assert.rejects(shim.query(handle, 'SELECT 1', []), /invalid connection handle/);
    });
}

test('failed SQL still closes the prepared statement', async (t) => {
    const { shim, calls } = harness(t, 'haybarn', { query: true });
    await shim.ready;
    const handle = await shim.open(':memory:');
    await assert.rejects(shim.query(handle, 'bad SQL', []), /query failed/);
    assert.deepEqual(calls.at(-1), ['statement-close']);
    await assert.rejects(shim.exec(handle, 'bad SQL', []), /query failed/);
    assert.deepEqual(calls.at(-1), ['statement-close']);
});

test('read-only file connections escape paths and use the attached database', async (t) => {
    const { shim, calls } = harness(t);
    await shim.ready;
    await shim.open("it's.duckdb?access_mode=read_only");
    assert.deepEqual(calls.filter((c) => c[0] === 'sql'), [
        ['sql', "ATTACH 'it''s.duckdb' AS attached (READ_ONLY)"], ['sql', 'USE attached'],
    ]);
});

test('failed ATTACH releases the connection', async (t) => {
    const { shim, calls } = harness(t, 'haybarn', { attach: true });
    await shim.ready;
    await assert.rejects(shim.open('missing.duckdb'), /attach failed/);
    assert.deepEqual(calls.at(-1), ['close']);
});

test('failed connection close can be retried', async (t) => {
    const failures = { close: true };
    const { shim } = harness(t, 'haybarn', failures);
    await shim.ready;
    const handle = await shim.open(':memory:');
    await assert.rejects(shim.close(handle), /close failed/);
    failures.close = false;
    await shim.close(handle);
    await assert.rejects(shim.query(handle, 'SELECT 1', []), /invalid connection handle/);
});

test('failed bootstrap terminates the worker and releases its URL', async (t) => {
    const { shim, calls } = harness(t, 'haybarn', { instantiate: true });
    await assert.rejects(shim.ready, /instantiate failed/);
    assert.deepEqual(calls.slice(-2), [['terminate'], ['revoke', 'blob:test-worker']]);
});

test('Haybarn exposes setup-time prefetch; DuckDB does not', async (t) => {
    const { shim, calls } = harness(t, 'haybarn');
    await shim.prefetchExtensions(['spatial', 'iceberg', 'avro', 'parquet']);
    assert.deepEqual(calls.at(-1), ['prefetch', ['spatial', 'iceberg', 'avro', 'parquet']]);
});

test('upstream DuckDB has no Haybarn-only prefetch method', async (t) => {
    const { shim } = harness(t);
    await shim.ready;
    assert.equal(shim.prefetchExtensions, undefined);
});

test('importing a second engine cannot replace active connections', async (t) => {
    const { shim, api, serialize } = harness(t);
    await shim.ready;
    assert.throws(() => installDuckDBBridge(api, serialize, { engine: 'haybarn', bundles: {} }), /already installed/);
    assert.equal(globalThis.ntshim.duckdb, shim);
});
