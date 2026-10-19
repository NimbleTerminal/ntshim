// Package pdfium is a ntshim that exposes upstream @embedpdf/pdfium
// (a PDFium WebAssembly build) to Go-WASM programs as a Document type
// that loads PDFs from memory and rasterizes pages to *image.RGBA.
//
// Under GOOS=js:
//
//	doc, err := pdfium.Load(ctx, pdfBytes)
//	defer doc.Close()
//	img, err := doc.RenderPage(ctx, 0, 144) // page 0 at 144 DPI
//
// Pages are 0-indexed; page sizes are PostScript points (1/72 inch).
//
// On any other GOOS, every public entry point returns ErrNonJSPlatform
// so host-side test binaries can import the package without panicking.
//
// The page must load ntshim/pdfium/pdfium-shim.js (via go tool
// ntshim-assets) before the WASM main runs; otherwise Load returns
// ErrBridgeMissing.
package pdfium
