# Audio browser assets

The default audio shim lives entirely in [`audio`](../../audio), using
`syscall/js` to schedule Web Audio `AudioBufferSourceNode`s. No JavaScript bridge,
CDN dependency, served worklet, or cross-origin isolation is needed.

`ntshim-assets <web-dir> --shim=audio` succeeds without writing any files or
creating directories, including in `--vendored` mode. There is currently no
worklet backend or worklet option. Browser-native Web Audio needs no upstream
version pin.
