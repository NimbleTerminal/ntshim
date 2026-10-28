# `ntshim` CHANGELOG

All notable changes to ntshim are documented in this file. Releases through
v0.5.0 below belong to the original booba-shim module and repository.

## `v0.6.1` - 2026-10-28

 * Raise the minimum Go version to 1.26.0.
 * Update Apache Arrow Go to v18.8.0 and refresh indirect Go dependencies.
 * Update CI to use `actions/checkout`, `actions/setup-go`, and
   `actions/setup-node` v7.

## `v0.6.0` - 2026-10-19

 * Rename booba-shim to ntshim, using the Go module `nimbleterminal.dev/ntshim`
   and repository [`NimbleTerminal/ntshim`](https://github.com/NimbleTerminal/ntshim).
   Update imports, use `go tool ntshim-assets`, and switch page assets to
   `ntshim/` and the browser namespace to `window.ntshim`. See the
   [migration guide](./README.md#migrating-from-booba-shim).
 * Update the vendored flint-ntcharts compiler to v0.3.1. The browser bundle
   is unchanged from v0.3.0; retain the local `ntshim` namespace patch.
 * Add optional Haybarn-Wasm 1.5.6-rc6 support alongside DuckDB-Wasm. Choose
   `haybarn-shim.js` or the default `duckdb-shim.js` with the same Go SQL and
   Arrow APIs. Haybarn supports setup-time extension prefetch; both database
   examples accept `?engine=haybarn`.

## `v0.5.0` - 2026-10-11

 * Add [`booba-shim/audio`](./audio) for mono/stereo int16 PCM playback through
   Web Audio. Includes browser resampling, configurable chunk size, lead time
   and gain, gesture unlock, backpressure, and draining pause/close.
 * Expose a portable `Player` interface and `NewSilent()` paced output.
   Non-browser `New` returns `ErrUnsupported`, allowing explicit fallback
   without native audio dependencies.
 * Accept `booba-shim-assets --shim=audio` without writing assets. The audio
   shim needs no page JavaScript or external browser library.
 * Add `examples/audio` with generated tones, gesture unlock, pause/resume,
   source-format changes, and a manual browser checklist.
 * Validation: host tests, audio race tests, and the full Go/WASM build pass.
   Safari 27.0.1 UI checks cover the example and the standard Booba page.
   Audible gaplessness and hidden-tab playback still need listening checks.

## `v0.4.0` - 2026-10-02

 * Update the vendored flint-ntcharts compiler to
   [v0.3.0](https://github.com/NimbleMarkets/flint-ntcharts/releases/tag/v0.3.0):
   an opt-in raster renderer. An input with `"renderer": "raster"` compiles to
   flint's ECharts option (`{"echarts", ...}` in place of `{"spec", ...}`) for a
   host to draw as an image; inputs without it compile exactly as before. The
   bundle grows from 233 KB to 734 KB. Read results with flint-ntcharts's
   `envelope.ParseResult`.

## `v0.3.0` - 2026-10-01

 * Update the vendored flint-ntcharts compiler to
   [v0.2.0](https://github.com/NimbleMarkets/flint-ntcharts/releases/tag/v0.2.0):
   new chart types (ECDF, connected scatter, bubble, histogram, area, lollipop,
   calendar heatmap), logarithmic axes, and charts that fill the size they are
   asked for. The compiled spec needs ntcharts v2.6.0 to render.

## `v0.2.0` - 2026-09-30

 * Added [`booba-shim/flintchart`](./flintchart): compile a flint
   `ChartAssemblyInput` (JSON) into an ntcharts render envelope. The
   compiler is a self-contained bundle vendored from
   [flint-ntcharts v0.1.0](https://github.com/NimbleMarkets/flint-ntcharts/releases/tag/v0.1.0)
   (`flint-chart@0.5.1`; see `web/flintchart/PROVENANCE.txt`), so nothing is
   fetched at run time.
   * `examples/flintchart-compile`: minimal browser example proving the
     compile bridge
 * Update DuckDB-Wasm from `1.29.0` to stable `1.32.0`, including the
   embedded workers and WASM artifacts. Keep Apache Arrow JS at `17.0.0`.
 * Update EmbedPDF PDFium from `2.14.2` to `2.15.1`, with matching JS and
   WASM CDN pins. Verified both upgrades in Chrome using the Go/WASM SQL,
   Arrow IPC, and PDF rendering examples; no bridge API changes required.

## `v0.1.0` - 2026-06-13

 * Initial release of `booba-shim`
   * [`booba-shim/duckdb`](./duckdb): DuckDB-Wasm via `database/sql` driver + native Arrow API 
   * [`booba-shim/pdfium`](./pdfium): PDF rasterization to `*image.RGBA` via PDFium-Wasm
