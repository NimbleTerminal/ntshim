# Embed Overlay: Interactive Web Widgets over the Terminal Grid

**Date:** 2026-10-01
**Status:** Design (awaiting review)
**Affected:** new `embed` package, new `web/` runtime, `cmd/ntshim-assets`, new example

## Goal

Let a Go-WASM TUI hosted by go-booba mount interactive web widgets (ECharts and
`<video>` first; WebGPU canvases and others later) over reserved cell regions.
Widget events reach the Bubble Tea model as messages, and the model drives the
widget's data, so the TUI stays the source of truth. Outside a browser, the
same `Embed` renders a caller-supplied fallback model.

**Success criterion:** one Go program shows an ECharts chart in the browser
whose click filters a TUI table, and shows an ntcharts/text rendering of the
same chart in a native terminal.

## Non-goals (v1)

- Kitty-graphics pixel backend (possible later; depends on the unpublished
  ghostty-web patch, see the archived GPU stack notes).
- Per-widget `Clip []Rect` occlusion; sub-cell sizing; widgets in scrollback.
- Frame-by-frame video fallback.

## Assumptions

- Lives in ntshim, following the `_js.go` / `_other.go` split.
- Cell geometry is derivable from `booba.terminal` (cols, rows, canvas size).
  **Unverified; step 0 of the plan is a spike with a go/no-go.** If it fails,
  the remedy is a small upstream hook in go-booba, not shim workarounds.

## 1. Go↔JS contract

```go
e := embed.New("chart1", "echarts",
    embed.WithProps(option),        // JSON, kind-specific
    embed.WithFallback(model))      // tea.Model for non-browser
e.View(w, h)                        // blank cells in browser, fallback.View() natively
// Update() receives embed.EventMsg{ID, Name, Data}
e.Send("seek", 12.5)                // Go -> widget command
```

- `View` is pure and returns a string, so rects are not published through it.
  The model reports rects through an explicit `embed.Layout` pass. After each
  frame the shim reconciles the reported set with JS: mount new, reposition
  moved, dispose absent.
- JS exposes `ntshim.embed.register(kind, adapter)` with
  `adapter = {mount(el, props), update(props), resize(w, h), command(name, arg),
  dispose(), on(event, cb)}`. The adapter registry lives in JS.
- Events travel JS→Go over one bridge channel, in the style of the duckdb shim.

## 2. Geometry, focus, occlusion

**Geometry.** One overlay container (`position:absolute`, `pointer-events:none`)
over the terminal canvas. Each widget is a child with `pointer-events:auto`.
Cell size = canvas size / (cols, rows), re-read on resize and font change
(ResizeObserver plus a terminal resize hook). Pixel rect = cell rect × cell
size, rounded to whole pixels; widgets snap to the grid.

**Focus.** Pointer events go to the widget inside its rect, to the terminal
elsewhere. Clicking a widget focuses it; keys then go to the widget only. An
escape chord (default `Esc Esc`, configurable) or a click outside returns focus
to the terminal. The shim emits `FocusMsg{ID, Focused}`.

**Occlusion.** A DOM overlay cannot be partially covered by cells. Each widget
has `Visible bool`; the model sets it false while a modal or popup overlaps.
Widgets belong to fixed layout regions; on scroll the model moves or hides them.

**Lifecycle.** Widgets are keyed by `ID`. Absent from a layout pass means
disposed: ECharts calls `chart.dispose()`; video pauses and releases its source.

## 3. Adapters, fallback, packaging, testing

**ECharts.** `props` is an ECharts `option`, or a flint `ChartAssemblyInput`
compiled to an `option` by the existing flintchart shim, so one spec drives
ECharts (browser) and ntcharts (fallback). Events `click`, `hover`,
`brushSelected` map to `EventMsg` with a serializable payload
`{seriesIndex, dataIndex, name, value}`; no DOM objects. Updates use
`setOption`; resize resizes the instance. ECharts is vendored with a
`PROVENANCE.txt` and loaded lazily.

**Video.** `props = {src, autoplay, loop, muted, controls}` on a native
`<video>`. Events: `play`, `pause`, `ended`, `timeupdate` (throttled ~4 Hz),
`error`, `blocked` (autoplay denied by browser gesture rules). Commands via
`Send("seek", t)` etc. Native fallback is a text line (title, duration, URL).

**Fallback.** `embed_other.go` never touches JS: `View` renders the fallback
model in the rect and no `EventMsg` is produced. In the browser, if the overlay
cannot initialize (metrics unreadable, adapter missing), the shim logs once and
renders the fallback in-grid rather than a blank rect.

**Packaging.** `embed/` with `embed_js.go`, `embed_other.go`, `doc.go`; JS in
`web/`; vendored copy under `cmd/ntshim-assets/assets/embed/`.
`ntshim-assets --shim=embed` copies the overlay runtime;
`--shim=embed-echarts` adds ECharts so video-only users do not pull it in.

**Testing.**
- Go: layout reconciliation (mount, move, dispose) against a fake bridge,
  runnable natively.
- JS: node/jsdom tests for the overlay container and adapter lifecycle;
  `smoke.mjs` in the example, as flintchart-compile does.
- Example `examples/embed-echarts/`: chart filters a table in the browser; the
  same program run natively demonstrates the fallback.

## Risks

1. Cell-metrics reliability (spike, go/no-go before adapters).
2. Keyboard capture while a widget is focused; mitigated by the escape chord.
3. Occlusion is Visible-only in v1; floating menus over a chart must hide it.
