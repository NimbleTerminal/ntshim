# Audio demo

Run `task serve` here, then open <http://localhost:8000>. Or build manually:

```sh
GOOS=js GOARCH=wasm go build -o app.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
python3 -m http.server 8000
```

The same `app.wasm` can be hosted with the standard Booba page. The program
creates its own controls from Go and needs no shim script or custom HTML IDs.
`go tool ntshim-assets . --shim=audio` is optional and writes nothing.

Manual browser checklist (host tests do not verify audible playback):

- Open a fresh page with autoplay blocked. Click the explanatory text or press a
  key. Hear a steady 440 Hz tone, with no gaps in a foreground tab.
- Pause while playing. The queued tail finishes, then silence. Resume continues.
- Select 16 kHz mono, then 44.1 kHz stereo. Both play at the correct pitch; the
  stereo tone has 440 Hz left and 660 Hz right. The browser resamples.
- Allow a full tone to finish. Its final partial chunk plays before suspension.
- Close during playback. The tail drains and further controls do not play audio.
- Reload and click Close before any other gesture. Pointerdown unlocks the
  context, the pending write completes, and Close drains. An API call to drain
  before a gesture must wait for unlock; wall-clock time cannot drain a stopped
  audio clock.
- Switch tabs for several seconds. Note possible gaps with the default 300 ms
  lead; compare with a longer `LeadTime` in `audio.Options`.

The status reports queued time, which can lead audible time. No browser playback
checks are implied by a successful Go build; record browser/version and results
when running this checklist.

Verification on 2026-10-11: host tests and the full `GOOS=js GOARCH=wasm` build
passed. Safari 27.0.1 loaded both this page and the unmodified go-booba v0.9.0
asset-tool template. The UI showed progress after gesture unlock, pause after
draining, resumed progress, mono/stereo format changes, completion after the
final partial chunk, and a successful Close while playing. Safari displayed
its playing-audio indicator during playback. Audible gaplessness, pitch, and
hidden-tab behavior still need a listening check; UI observations do not verify
those properties.
