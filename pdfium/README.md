# ntshim/pdfium

Render PDFs in Go-WASM programs via [@embedpdf/pdfium](https://www.embedpdf.com/docs/pdfium)
(a PDFium WebAssembly build). Loads documents from memory and
rasterizes pages to `*image.RGBA`.

## Quickstart

**main.go** (GOOS=js GOARCH=wasm):

```go
doc, err := pdfium.Load(ctx, pdfBytes)
if err != nil { /* errors.Is(err, pdfium.ErrBridgeMissing) ⇒ no shim on page */ }
defer doc.Close()

n, _ := doc.NumPages()                  // page count
w, h, _ := doc.PageSize(0)              // points (1/72 inch), 0-indexed
img, err := doc.RenderPage(ctx, 0, 144) // *image.RGBA at 144 DPI
```

**index.html** (excerpt):

```html
<script src="wasm_exec.js"></script>
<script type="module">
  import './ntshim/pdfium/pdfium-shim.js';
  const go = new Go();
  const wasm = await WebAssembly.instantiateStreaming(fetch('app.wasm'), go.importObject);
  go.run(wasm.instance);
</script>
```

The Go side awaits the bridge's `ready` promise on first `Load`, so the
page does not need to await it (unlike ntshim/duckdb).

**Assets:** `go tool ntshim-assets <web-dir> --shim=pdfium`

## Semantics

- Pages are 0-indexed. Page sizes are PostScript points.
- `Document` methods are safe for concurrent use; rendering and Close
  are serialized internally.
- `Close` is idempotent. A `Document` dropped without `Close` is
  reclaimed by a GC cleanup, but call `Close` — the PDF bytes stay
  pinned in the PDFium heap until then.
- Under GOOS != js every entry point returns `ErrNonJSPlatform`.
- DPI limits / pixel budgets are the caller's policy; PDFium's wasm
  heap is fixed-size, so very large renders (e.g. Letter @ 300 DPI ≈
  34 MB) can fail — catch the error or clamp DPI using `PageSize`.

## Pinned upstream version

- `@embedpdf/pdfium@2.15.1` — bump in `web/pdfium/pdfium-shim.js`.
