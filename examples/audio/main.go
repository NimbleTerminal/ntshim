//go:build js && wasm

package main

import (
	"fmt"
	"math"
	"syscall/js"

	"nimbleterminal.dev/ntshim/audio"
)

// The controls are created by Go, so this also runs on an unmodified Booba
// page. No audio script, HTML IDs, or Booba-specific imports are required.
func main() {
	doc := js.Global().Get("document")
	panel := doc.Call("createElement", "section")
	panel.Get("style").Set("cssText", "padding:16px;background:#fff;color:#111;font:16px sans-serif;position:relative;z-index:1000")
	doc.Get("body").Call("appendChild", panel)
	title := doc.Call("createElement", "p")
	title.Set("textContent", "Audio demo: click here or press a key to unlock sound. Each tone lasts 10 seconds. Pause drains queued sound; resume continues.")
	panel.Call("appendChild", title)
	status := doc.Call("createElement", "p")
	panel.Call("appendChild", status)
	gain := 0.15
	p, err := audio.New(audio.Options{Gain: &gain})
	if err != nil {
		status.Set("textContent", err.Error())
		select {}
	}
	commands := make(chan string, 16)
	for _, label := range []string{"Pause", "Resume", "16 kHz mono", "44.1 kHz stereo", "Close"} {
		button := doc.Call("createElement", "button")
		button.Set("textContent", label)
		fn := js.FuncOf(func(js.Value, []js.Value) any {
			select {
			case commands <- label:
			default:
			}
			return nil
		})
		button.Call("addEventListener", "click", fn)
		panel.Call("appendChild", button)
	}
	go func() {
		rate, channels, frame := 16000, 1, 0
		paused := false
		if err := p.SetSampleRate(rate, channels); err != nil {
			status.Set("textContent", err.Error())
			return
		}
		p.Enable(true)
		status.Set("textContent", "Waiting for gesture: 16 kHz mono, 440 Hz tone")
		for {
			var cmd string
			if paused {
				cmd = <-commands
			} else {
				select {
				case cmd = <-commands:
				default:
				}
			}
			switch cmd {
			case "Pause":
				p.Enable(false)
				paused = true
				status.Set("textContent", "Paused after draining")
			case "Resume":
				p.Enable(true)
				paused = false
			case "16 kHz mono", "44.1 kHz stereo":
				p.Enable(false)
				rate, channels, frame = 16000, 1, 0
				if cmd == "44.1 kHz stereo" {
					rate, channels = 44100, 2
				}
				if err := p.SetSampleRate(rate, channels); err != nil {
					status.Set("textContent", err.Error())
					return
				}
				p.Enable(true)
				paused = false
			case "Close":
				err := p.Close()
				status.Set("textContent", fmt.Sprintf("Closed (drained): %v", err))
				return
			}
			if paused {
				continue
			}
			// An uneven ending exercises flushing a partial chunk.
			count := min(256, rate*10+123-frame)
			if count == 0 {
				p.Enable(false)
				paused = true
				status.Set("textContent", "Tone finished and drained. Choose a tone to play again.")
				continue
			}
			samples := make([]int16, count*channels)
			for i := range count {
				for ch := range channels {
					frequency := 440.0 + float64(ch)*220
					samples[i*channels+ch] = int16(32767 * math.Sin(2*math.Pi*frequency*float64(frame+i)/float64(rate)))
				}
			}
			n, err := p.Write(samples)
			frame += n / channels
			if err != nil {
				status.Set("textContent", err.Error())
				return
			}
			status.Set("textContent", fmt.Sprintf("%d Hz, %d channel(s): %.2f seconds queued (up to the lead time ahead of sound)", rate, channels, float64(frame)/float64(rate)))
		}
	}()
	select {}
}
