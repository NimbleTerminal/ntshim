# ntshim/flintchart

Compile a flint `ChartAssemblyInput` (JSON) into an ntcharts render
envelope in Go-WASM programs, via the vendored
[flint-ntcharts](https://github.com/NimbleMarkets/flint-ntcharts) browser
bundle.

## Quickstart

**main.go** (GOOS=js GOARCH=wasm):

```go
out, err := flintchart.Compile(ctx, inputJSON)
if err != nil { /* errors.Is(err, flintchart.ErrBridgeMissing) ⇒ no shim on page */ }
// out is envelope JSON: {"spec",...} on success, {"error":{"message"}}
// on a compile failure — the latter is data, not err. Parse it (e.g.
// with flint-ntcharts's envelope.ParseResult) to distinguish the two.
// With "renderer": "raster" in the input, success is {"echarts",...}
// (flint's ECharts option) instead of {"spec",...}.

v, _ := flintchart.Version() // "flint-ntcharts flint-chart@0.5.1"
```

**index.html** (excerpt):

```html
<script src="wasm_exec.js"></script>
<script type="module">
  import './ntshim/flintchart/flintchart-shim.js';
  const go = new Go();
  const wasm = await WebAssembly.instantiateStreaming(fetch('app.wasm'), go.importObject);
  go.run(wasm.instance);
</script>
```

The bridge is a fully self-contained bundle with no wasm/network
dependency — its `ready` promise is already resolved by the time it
loads, but the Go side awaits it anyway on first `Compile`/`Version`
call, for symmetry with pdfium/duckdb.

**Assets:** `go tool ntshim-assets <web-dir> --shim=flintchart`

**Supported chart types:** whatever the vendored flint-ntcharts bundle
supports — currently all 7: Bar Chart, Stacked Bar Chart, Line Chart,
Scatter Plot, Heatmap, Candlestick Chart, and Sparkline (see
flint-ntcharts's README "Supported chart types" for the authoritative,
up-to-date list — this bridge has no chart-type logic of its own, it
only forwards to the bundle).

## Semantics

- `Compile` is stateless and synchronous under the hood; every call is
  independent (no handle to manage or Close).
- Compile failures reported by the compiler are returned as a normal
  `[]byte` envelope (`{"error":{"message"}}`), not a Go error. Only a
  JS-level throw (broken bundle, OOM) surfaces as a `*flintchart.Error`.
- Under GOOS != js every entry point returns `ErrNonJSPlatform`.

## Limitations

- `Compile` is a synchronous call under the hood (no JS event-loop turn):
  budget roughly 10-30ms for a typical chart spec. That's fine for a
  one-shot compile off the render path (e.g. reacting to a data/spec
  change), but don't call it in a tight per-frame loop.
- Browser-only. `GOOS=js` is required; everything returns
  `ErrNonJSPlatform` elsewhere (see Semantics above).
- Input and output both cross the boundary as UTF-8 JSON strings — no
  binary/Uint8Array marshaling, no wasm module, no handle to manage.

## Example

See [`examples/flintchart-compile`](../examples/flintchart-compile) for a
minimal end-to-end program: it compiles a small flint input in the
browser and displays the envelope's size, warning count, and
pretty-printed spec JSON.

For turning the envelope into an ntcharts render (bar/line charts in a
Bubble Tea TUI), see
[flint-ntcharts](https://github.com/NimbleMarkets/flint-ntcharts)'s
`envelope.Parse` and the rest of its Go toolchain — this package only
gets you the compiled spec JSON; rendering it is out of scope here.

## Using with flint-ntcharts's TUI

flint-ntcharts's `tui` package renders through a `tui.Compiler` interface,
not this package's `Compile` directly, so it can run against either the
native embedded-wasm backend or this browser bridge. Satisfying it is a
thin adapter that applies `envelope.Option`s to the input, calls this
package's `Compile`, and parses the resulting envelope:

```go
type Adapter struct{}

func (Adapter) Compile(ctx context.Context, input []byte, opts ...envelope.Option) (spec.Spec, []envelope.Warning, error) {
	input, err := envelope.Apply(input, opts)
	if err != nil {
		return spec.Spec{}, nil, err
	}
	out, err := flintchart.Compile(ctx, input)
	if err != nil {
		return spec.Spec{}, nil, err
	}
	return envelope.Parse(out)
}
```

**Implementations MUST apply `opts` to the input (`envelope.Apply`) before
compiling.** Dropping them still compiles fine — there's no error, no
warning — but the chart is compiled against its original, unadjusted size
and so silently never fits the window: the TUI's resize/fit-to-window
behavior depends entirely on `opts` (e.g. `envelope.WithBaseSize`)
reaching the compiler on every call.

## Vendored bundle

`flintchart-shim.js` is vendored from flint-ntcharts
(`js/dist/flintchart-shim.mjs`) — see `web/flintchart/PROVENANCE.txt`
for the source commit and sha256. Re-vendor by running `make
browser-shim` in flint-ntcharts, applying the namespace patch described in
`web/flintchart/README.md`, and copying the result over both
`web/flintchart/` and `cmd/ntshim-assets/assets/flintchart/`.
