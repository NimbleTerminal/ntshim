# ntshim/flintchart JS bridge

This directory ships in the asset tool. End users do not edit these files
directly — they run `go tool ntshim-assets <web-dir> --shim=flintchart`,
which writes them to `<web-dir>/ntshim/flintchart/`.

Unlike `pdfium` and `duckdb`, this bridge is not built against a pinned
upstream npm package fetched at asset-install time — the ntcharts backend
for flint-chart lives in flint-ntcharts and is not published to npm, so
there is no CDN URL to point at. `flintchart-shim.js` is instead a single
self-contained bundle vendored from a tagged release of
[flint-ntcharts](https://github.com/NimbleMarkets/flint-ntcharts)
(`js/dist/flintchart-shim.mjs`, built by `make browser-shim` there), with the
browser entry point patched from `globalThis.boobaShim` to `globalThis.ntshim`. It has
no runtime imports (no CDN fetch, no wasm) — loading the file synchronously
sets `window.ntshim.flintchart` (aka `globalThis.ntshim.flintchart`)
to an object with `compile`, `version`, and an already-resolved `ready`
promise.

`PROVENANCE.txt` records the flint-ntcharts tag and commit, the flint-chart
version, upstream sha256, local namespace patch, and the shipped file's sha256
and size; `go test ./flintchart/` checks the
bundle against it and that both vendored copies are identical.

To re-vendor, build from a clean clone of the flint-ntcharts release tag —
never from a working tree with local edits:

```sh
git clone --branch vX.Y.Z https://github.com/NimbleMarkets/flint-ntcharts.git
cd flint-ntcharts && (cd js && npm ci) && make browser-shim
```

Record the upstream checksum before editing. Replace `globalThis.boobaShim`
with `globalThis.ntshim` in the browser entry point if upstream still uses
the old namespace. Copy the patched bundle over `flintchart-shim.js` here and
over the mirror under `cmd/ntshim-assets/assets/flintchart/`. Update both
`PROVENANCE.txt` files with the source tag and commit, upstream checksum,
patch description, and the patched bundle's sha256 and byte count.
