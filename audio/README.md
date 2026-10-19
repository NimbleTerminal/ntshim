# audio

PCM playback through Web Audio for Go `js/wasm`, with no page JavaScript or
external dependency.

```go
import "nimbleterminal.dev/ntshim/audio"

p, err := audio.New(audio.Options{})
if err != nil {
    p = audio.NewSilent()
}
defer p.Close()
if err := p.SetSampleRate(16000, 1); err != nil {
    return err
}
p.Enable(true)
// Run on a goroutine, not directly in a syscall/js event callback.
_, err = p.Write(samples) // []int16, interleaved; returns samples, not frames
p.Enable(false)          // flush the final partial chunk and drain
```

`Player` exposes `SetSampleRate(hz, channels) error`, `Write([]int16) (int, error)`,
`Enable(bool)`, and `Close() error`. `New(Options)` returns `(Player, error)` and
starts disabled. Call `SetSampleRate` and `Enable(true)` before writing. Mono and
stereo are supported; the browser validates the source rate and resamples each
buffer to its output rate. Changing format flushes partial PCM at the old format.
Malformed interleaved frames are rejected before consuming any samples.

`Options.LeadTime` defaults to 300 ms and `Options.ChunkFrames` to 2048. Zero
selects these defaults. `Options.Gain` is an optional `*float64`: nil is unity,
zero mutes, and positive finite values scale the output through a GainNode.
Use a gain below one for quieter generated tones.

Each channel is encoded as little endian float32 PCM and copied with one
`CopyBytesToJS` call per chunk. Chunks are scheduled back to back on the audio
clock, with 50 ms startup/recovery slack. Write blocks when scheduled audio is
more than the lead time ahead. Even a large Write is processed in bounded
chunks. A chunk can temporarily exceed the lead limit, and a partial chunk is
held until another Write fills it or Enable(false)/Close flushes it. Chunk size
therefore affects startup latency too (2048 frames are 128 ms at 16 kHz).
The producer's progress counts frames queued, not frames heard, so progress
normally runs up to the lead time ahead, with additional partial-chunk buffering.

Browsers require a user gesture to unlock sound. The package installs pointerdown
and keydown listeners, resumes on the first gesture while enabled, and removes
both listeners. Enable(true) also requests resume. Before unlock, Write can block
at the lead limit without accumulating unlimited audio. Enable(false) interrupts
a blocked Write (which reports the samples accepted and ErrDisabled), flushes,
drains queued audio, then suspends. Draining follows the audio clock, so a
drain before autoplay unlock waits for a gesture too. Close interrupts Write
with ErrClosed, drains queued audio, and releases listeners and the context.
Close is idempotent. Serialize control calls; they may run concurrently with
Write. After an interrupted Write, retry only the unconsumed suffix.

Hidden tabs often throttle Go timers to about once per second while the audio
clock continues. A small lead can cause gaps; a larger lead improves background
playback but increases pause/skip latency. 300 ms is a compromise, not a guarantee
of uninterrupted background audio. Pause drains queued sound rather than cutting
it off, so it also takes up to the queued duration. Use Enable(false) at end of
stream or Close while running to preserve the tail.

The interface, options, and errors are available on every platform. Outside
`js/wasm`, `New` returns `nil, ErrUnsupported`. In a browser without AudioContext,
its error also matches `errors.Is(err, audio.ErrUnsupported)`. Construction never
silently selects a fallback; the caller makes that choice, as above.

`NewSilent()` returns a portable `Player` that discards PCM and blocks for its
duration, keeping progress and animations in real time. It starts disabled and
requires a positive sample rate and one or two channels. It accepts samples in
small batches and drains the accepted batch on pause or close. A blocked Write
returns its accepted sample count and ErrDisabled or ErrClosed when interrupted.
After resume, retry only the unconsumed suffix. Pause time does not become a
playback backlog, and sample rate changes apply to subsequent writes.

The silent output has no configurable lead, gain, or chunk size. It accepts any
positive source rate and paces complete frames using the wall clock; no audio
device or browser is needed. Native desktop and hardware output can implement
`Player` in the application; this package does not include those backends.

Encoder, silent pacing, lifecycle, and unsupported-platform tests run on the host.
See the [browser example and manual checks](../examples/audio/README.md). No
worklet is shipped; `--shim=audio` is an asset-tool no-op.
