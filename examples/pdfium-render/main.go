//go:build js

package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"image/png"
	"syscall/js"

	"nimbleterminal.dev/ntshim/pdfium"
)

//go:embed testdata/Example.pdf
var examplePDF []byte

func main() {
	out := func(s string) {
		div := js.Global().Get("document").Call("getElementById", "out")
		div.Set("innerHTML", div.Get("innerHTML").String()+s+"<br/>")
	}

	ctx := context.Background()
	doc, err := pdfium.Load(ctx, examplePDF)
	if err != nil {
		out("ERROR load: " + err.Error())
		return
	}
	defer doc.Close()

	n, err := doc.NumPages()
	if err != nil {
		out("ERROR numPages: " + err.Error())
		return
	}
	w, h, err := doc.PageSize(0)
	if err != nil {
		out("ERROR pageSize: " + err.Error())
		return
	}
	out(fmt.Sprintf("pages=%d page0=%.1fx%.1fpt", n, w, h))

	img, err := doc.RenderPage(ctx, 0, 144)
	if err != nil {
		out("ERROR render: " + err.Error())
		return
	}
	out(fmt.Sprintf("rendered %dx%dpx", img.Rect.Dx(), img.Rect.Dy()))

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		out("ERROR png: " + err.Error())
		return
	}
	js.Global().Get("document").Call("getElementById", "page").
		Set("src", "data:image/png;base64,"+base64.StdEncoding.EncodeToString(buf.Bytes()))
	select {}
}
