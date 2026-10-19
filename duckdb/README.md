# ntshim/duckdb

A `database/sql` driver named `"duckdb"` plus a native Arrow API, both wrapping DuckDB-Wasm running in a browser Web Worker.

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
- `uint64`/`UBIGINT` values above `math.MaxInt64` return an error rather than wrapping silently (`driver.Value` cannot carry `uint64`); lossless large-`UBIGINT` support is a v0.2 task.
