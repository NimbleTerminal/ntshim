//go:build js

// Example flintchart-compile: compile a flint chart spec in the browser via
// the ntshim flintchart bridge and display the resulting ntcharts-spec
// envelope. Rendering the spec with ntcharts in a browser TUI is go-booba's
// domain; this example proves the compile bridge alone.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"

	"nimbleterminal.dev/ntshim/flintchart"
)

const sampleInput = `{
  "data": {"values": [
    {"month": "Jan", "revenue": 120}, {"month": "Feb", "revenue": 180},
    {"month": "Mar", "revenue": 95},  {"month": "Apr", "revenue": 210}
  ]},
  "semantic_types": {"revenue": {"semanticType": "Price", "unit": "USD"}},
  "chart_spec": {
    "chartType": "Bar Chart",
    "encodings": {"x": {"field": "month"}, "y": {"field": "revenue"}},
    "baseSize": {"width": 60, "height": 16}
  }
}`

func setText(id, text string) {
	doc := js.Global().Get("document")
	el := doc.Call("getElementById", id)
	if !el.IsUndefined() && !el.IsNull() {
		el.Set("textContent", text)
	}
}

func main() {
	raw, err := flintchart.Compile(context.Background(), []byte(sampleInput))
	if err != nil {
		setText("status", "compile bridge error: "+err.Error())
		return
	}
	var env struct {
		Spec     json.RawMessage `json:"spec"`
		Warnings []any           `json:"warnings"`
		Size     struct{ Width, Height int }
		Error    *struct{ Message string } `json:"error"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		setText("status", "bad envelope: "+err.Error())
		return
	}
	if env.Error != nil {
		setText("status", "compiler error: "+env.Error.Message)
		return
	}
	var pretty json.RawMessage = env.Spec
	out, _ := json.MarshalIndent(pretty, "", "  ")
	setText("status", fmt.Sprintf("compiled ok — %dx%d cells, %d warning(s)", env.Size.Width, env.Size.Height, len(env.Warnings)))
	setText("envelope", string(out))
	select {} // keep the Go runtime alive
}
