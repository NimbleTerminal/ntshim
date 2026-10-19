# ntshim

Browser shims that let Go-WASM TUIs hosted by [go-booba](https://github.com/NimbleMarkets/go-booba) access browser-only capabilities. Current shims cover DuckDB-Wasm, PDFium-Wasm, flint-chart, and Web Audio.

## Shims

| Subpackage | What it does |
|---|---|
| [`ntshim/duckdb`](./duckdb) | DuckDB-Wasm or Haybarn-Wasm via `database/sql` driver + native Arrow API |
| [`ntshim/pdfium`](./pdfium) | PDF rasterization to `*image.RGBA` via PDFium-Wasm |
| [`ntshim/flintchart`](./flintchart) | Compile a flint `ChartAssemblyInput` into an ntcharts render envelope via a vendored flint-chart bundle |
| [`ntshim/audio`](./audio) | Play mono/stereo int16 PCM through Web Audio, with resampling and backpressure |

## Install

With Go 1.25 or newer:

```sh
go get nimbleterminal.dev/ntshim@v0.6.0
go mod edit -tool=nimbleterminal.dev/ntshim/cmd/ntshim-assets
```

Then run `go tool ntshim-assets <web-dir> --shim=duckdb`
(or `--shim=pdfium` or `--shim=flintchart`) to populate your HTML directory.
Audio needs no browser assets; `--shim=audio` is accepted and writes nothing.
For Haybarn, use the same `--shim=duckdb` assets and import
`ntshim/duckdb/haybarn-shim.js` instead of `duckdb-shim.js` in the page.
See [engine selection and extension prefetch](./duckdb/README.md#choose-an-engine).

## Migrating from booba-shim

ntshim uses the Go module path `nimbleterminal.dev/ntshim` and lives in a new
repository at
[`github.com/NimbleTerminal/ntshim`](https://github.com/NimbleTerminal/ntshim).
The original booba-shim repository and its releases remain at their original
module path. This migration does not rely on a GitHub repository redirect.

Update these names together, then rebuild your WASM and regenerate the page assets:

| Before | After |
|---|---|
| `github.com/NimbleMarkets/booba-shim` (and subpackages) | `nimbleterminal.dev/ntshim` (same subpackages) |
| Root Go package `boobashim` | `ntshim` |
| `cmd/booba-shim-assets` / `go tool booba-shim-assets` | `cmd/ntshim-assets` / `go tool ntshim-assets` |
| Page asset directory `booba-shim/<shim>/` | `ntshim/<shim>/` |
| JavaScript `window.boobaShim` / `globalThis.boobaShim` | `window.ntshim` / `globalThis.ntshim` |
| TypeScript `BoobaShim`, `BoobaShimDuckDB`, `BoobaShimPDFium`, `BoobaShimFlintchart` | `NTShim`, `NTShimDuckDB`, `NTShimPDFium`, `NTShimFlintchart` |

Replace the old tool directive and imports before running the install commands
above and `go mod tidy`. Update HTML imports, build/clean tasks, and ignore rules
to the new asset directory. The Go shim APIs and bridge methods are unchanged;
the old browser namespace and command are not aliases. Deploy the rebuilt WASM
and updated page assets together. The versions through v0.5.0 in the changelog
describe releases of booba-shim.

## Quickstart — minimal browser SQL demo

The smallest working example using `ntshim/duckdb` via `database/sql`:

**main.go** (GOOS=js GOARCH=wasm):

```go
package main

import (
	"database/sql"
	"fmt"
	"syscall/js"

	_ "nimbleterminal.dev/ntshim/duckdb"
)

func main() {
	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	if _, err := db.Exec(`CREATE TABLE t (id INT64, name VARCHAR)`); err != nil {
		panic(err)
	}

	var id int64
	var name string
	if err := db.QueryRow(`SELECT 1 as id, 'hello' as name`).Scan(&id, &name); err != nil {
		panic(err)
	}

	js.Global().Call("alert", fmt.Sprintf("Got: %d, %s", id, name))
	select {}
}
```

**index.html** (excerpt):

```html
<div id="output"></div>
<script src="wasm_exec.js"></script>
<script type="module">
  import './ntshim/duckdb/duckdb-shim.js';
  await window.ntshim.duckdb.ready;
  const go = new Go();
  const wasm = await WebAssembly.instantiateStreaming(fetch('app.wasm'), go.importObject);
  go.run(wasm.instance);
</script>
```

**Taskfile.yml** (excerpt):

```yaml
tasks:
  build:
    cmds:
      - GOOS=js GOARCH=wasm go build -o app.wasm .
      - cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
      - go tool ntshim-assets . --shim=duckdb
```

## Asset tool

```sh
go tool ntshim-assets <web-dir> --shim=<name> [--shim=<name> ...] [--cdn|--vendored]
```

Writes each selected shim under `<web-dir>/ntshim/<shim>/`. `--shim=audio` succeeds without writing anything: audio needs no page JavaScript. Use `--cdn` (default) to load upstream browser dependencies from CDN. DuckDB also has a `--vendored` mode that embeds local bundles (see below).

### Vendored mode

The DuckDB `--vendored` bridge remains a stub — it embeds the upstream bundles but throws an explanatory error on every call. Use `--cdn` (the default) for production. See `web/duckdb/duckdb-shim-vendored.js` for the technical details and the path to a real implementation in a future release.

### Pinned upstream versions

- `@duckdb/duckdb-wasm@1.32.0` (duckdb shim; tracked in `web/duckdb/duckdb-shim.js`)
- `@haybarn/haybarn-wasm@1.5.6-rc6` (optional DuckDB-compatible engine; tracked in `web/duckdb/haybarn-shim.js`)
- `apache-arrow@17.0.0` (duckdb shim)
- `@embedpdf/pdfium@2.15.1` (pdfium shim; tracked in `web/pdfium/pdfium-shim.js`)
- `flint-ntcharts@0.3.1` with `flint-chart@0.5.1` (flintchart shim; vendored, not CDN-fetched — tracked in `web/flintchart/PROVENANCE.txt`)

## Audio

Added in booba-shim v0.5.0. See the [changelog](./CHANGELOG.md) for release details.

[`audio.New(audio.Options{})`](./audio/README.md) creates a Go `js/wasm` PCM
player with `SetSampleRate`, `Write`, `Enable`, and `Close`. It works on the
standard Booba page without a JavaScript bridge. Missing Web Audio returns a
constructor error so callers can choose a silent paced output. No worklet or
upstream browser dependency is needed.

The `Player` interface is portable. Outside `js/wasm`, `audio.New` returns
`audio.ErrUnsupported`; `audio.NewSilent()` supplies a paced silent player on
any platform. Applications can also implement `Player` with their native audio
backend. Fallback selection is explicit.

Call `SetSampleRate(hz, channels)` (mono or stereo), then `Enable(true)` and
write interleaved `[]int16` from a goroutine. A first click or key press unlocks
sound under browser autoplay rules. The browser resamples to the output device.
`Enable(false)` flushes the partial chunk, drains running sound, then suspends;
`Enable(true)` resumes. `Close` drains running audio and releases resources.

The defaults are 2048-frame chunks and a 300 ms lead, with configurable gain.
Write blocks for backpressure so producers stay near real time. Producer
progress counts frames queued, not frames heard, and runs up to the lead time
ahead, plus any partial chunk still being collected. A larger lead reduces gaps
but increases pause/skip latency. Hidden tabs can throttle Go timers to roughly
once per second while the audio clock keeps running, so the default lead can
produce background gaps. See the [tone demo](./examples/audio) for gesture
unlock, pause/resume, and manual browser checks.

## License

[MIT](./LICENSE.txt) — Copyright (c) 2026 Neomantra Corp.
