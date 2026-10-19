# ntshim/duckdb

A `database/sql` driver named `"duckdb"` plus a native Arrow API, wrapping either
DuckDB-Wasm or Haybarn-Wasm in a browser Web Worker.

## Choose an engine

Install the bridge assets with `go tool ntshim-assets <web-dir> --shim=duckdb`.
The page chooses one engine entry point before starting Go/WASM:

```javascript
// Upstream DuckDB-Wasm (default):
import './ntshim/duckdb/duckdb-shim.js';
await window.ntshim.duckdb.ready;
```

For Haybarn, replace that import with:

```javascript
import './ntshim/duckdb/haybarn-shim.js';
await window.ntshim.duckdb.ready;
```

Both install `window.ntshim.duckdb`. The Go import, `sql.Open("duckdb", dsn)`,
and native `duckdb.Open` / `QueryArrow` APIs stay the same. Only the selected
engine is downloaded. Load one entry point per page; importing both is an error.
`window.ntshim.duckdb.engine` reports `"duckdb"` or `"haybarn"`.

DuckDB-Wasm is pinned to **1.32.0**; Haybarn-Wasm to **1.5.6-rc6**, a release
candidate. Haybarn uses its own extension repository and signing keys. Its
small JavaScript module is shipped with the bridge; worker and WASM files load
from unpkg, since its package exceeds jsDelivr's size limit. Both engines use
Apache Arrow **17.0.0** for browser queries and IPC serialization.
Both entry points select the single-threaded MVP or exception-handling bundle
according to browser support.

The [SQL](../examples/duckdb-sql) and [Arrow](../examples/duckdb-arrow) examples
run with either engine: open the page with `?engine=haybarn` to try Haybarn,
or `?engine=duckdb` for the default.

### Haybarn extension prefetch

Haybarn optionally downloads, hashes and compiles extensions ahead of `LOAD`:

```javascript
import './ntshim/duckdb/haybarn-shim.js';
await window.ntshim.duckdb.ready;
await window.ntshim.duckdb.prefetchExtensions(['spatial']);
// Start Go/WASM. Its SQL can now run LOAD spatial.
```

Prefetch does not load extensions. Call it once during setup, before launching
the application's queries, and after any `INSTALL ... FROM community` or
repository settings that affect extension URLs. Include transitive extension
dependencies (for example, `iceberg` also needs `avro` and `parquet`). Downloads
that fail fall back to normal `LOAD` behavior. This method is available only
on the Haybarn entry point; DuckDB keeps its normal extension-loading behavior.
See [Query.Farm's startup analysis](https://query.farm/blog/duckdb-wasm-startup-time/)
for the upstream behavior and measurements.

## Validation

Verified on 2026-10-19 in Chrome using the actual Go/WASM SQL and Arrow
examples with both engines. Browser checks also passed for parameterized
Arrow IPC queries, file-buffer and file-URL registration, and Haybarn's
`prefetchExtensions(['autocomplete'])` followed by `LOAD autocomplete` and
an extension query. Bridge lifecycle and error handling are covered by
`node --test web/duckdb/bridge.test.mjs`.

## DSN format

```
<path>?key=val&key=val…
```

Where `<path>` is `:memory:` or a virtual filename pre-registered via the JS bridge. Supported query parameters:

- `access_mode=read_only` — enable read-only mode.

Example: `"my-file.db?access_mode=read_only"`.

## File registration

Before calling `sql.Open`, the page (or Go-side code via `syscall/js`) must pre-register files with the bridge.

**From JavaScript:**

```javascript
// Register a file from a URL (lazy-loaded on first access)
await ntshim.duckdb.registerFileURL('my-file.db', '/path/to/file.db');

// Register a file from bytes (immediate)
const bytes = new Uint8Array([/* ... */]);
await ntshim.duckdb.registerFileBuffer('data.csv', bytes);
```

**From Go (inside a WASM module):**

```go
import "syscall/js"

js.Global().Get("ntshim").Get("duckdb").Call("registerFileURL", "my-file.db", "/path/to/file.db")
```

Then open the connection:

```go
db, _ := sql.Open("duckdb", "my-file.db")
```

## Supported types (v0.1.0)

**Argument types** (passed to `db.QueryRow`, `db.Query`, `db.Exec`):

- `nil` → NULL
- `bool` → BOOLEAN
- signed integers (`int`, `int8`, `int16`, `int32`, `int64`) → BIGINT
- unsigned integers (`uint`, `uint8`, `uint16`, `uint32`, `uint64`) up to `math.MaxInt64` → BIGINT
- `float32`, `float64` → DOUBLE
- `string` → VARCHAR
- `[]byte` → BLOB
- `time.Time` → TIMESTAMP (ISO-8601 format, UTC)

**Arrow column types** (returned by `db.Query` / `db.QueryRow.Scan`):

- Boolean
- Int8, Int16, Int32, Int64
- Uint8, Uint16, Uint32, Uint64
- Float32, Float64
- String, LargeString
- Binary, LargeBinary
- Date32, Date64
- Timestamp

Anything else errors loudly; no silent truncation or zero-value fallback.

## Async / goroutine safety

Every method blocks the calling goroutine on a JS Promise. This is safe inside `tea.Cmd` goroutines (which run off the main thread); it is **not safe** on Bubble Tea's main `Update` goroutine.

**Single-connection concurrency:** callers must serialize all operations on a single `*sql.DB`. Create a mutex or use a channel if you need true concurrency.

## Known limitations (v0.1.0)

- **Only one file-path connection per page session.** The bridge hardcodes the ATTACH alias `attached`, so opening a second `sql.Open("duckdb", "other.duckdb?...")` in the same page session fails with a catalog error. `:memory:` connections are unaffected. Per-handle aliases are a v0.2 task.
- `rowsAffected` is always 0; DuckDB-Wasm does not expose row-count data.
- No transactions.
- No streaming `send()` path; all results are fully materialized via `query()` before returning.
- `context.Context` is accepted but cancellation is not implemented.
- `--vendored` asset-tool mode is a stub; use `--cdn` for production.
- Haybarn is CDN-only; `--vendored` does not install its entry point.
- `uint64`/`UBIGINT` values above `math.MaxInt64` return an error rather than wrapping silently (`driver.Value` cannot carry `uint64`); lossless large-`UBIGINT` support is a v0.2 task.
