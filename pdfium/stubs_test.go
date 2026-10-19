//go:build !js

package pdfium_test

import (
	"context"
	"errors"
	"testing"

	"nimbleterminal.dev/ntshim/pdfium"
)

func TestLoadReturnsErrNonJSPlatform(t *testing.T) {
	_, err := pdfium.Load(context.Background(), []byte("%PDF-1.4"))
	if !errors.Is(err, pdfium.ErrNonJSPlatform) {
		t.Fatalf("Load err = %v, want ErrNonJSPlatform", err)
	}
}

func TestDocumentMethodsReturnErrNonJSPlatform(t *testing.T) {
	var d pdfium.Document
	if _, err := d.NumPages(); !errors.Is(err, pdfium.ErrNonJSPlatform) {
		t.Fatalf("NumPages err = %v, want ErrNonJSPlatform", err)
	}
	if _, _, err := d.PageSize(0); !errors.Is(err, pdfium.ErrNonJSPlatform) {
		t.Fatalf("PageSize err = %v, want ErrNonJSPlatform", err)
	}
	if _, err := d.RenderPage(context.Background(), 0, 72); !errors.Is(err, pdfium.ErrNonJSPlatform) {
		t.Fatalf("RenderPage err = %v, want ErrNonJSPlatform", err)
	}
	if err := d.Close(); err != nil {
		t.Fatalf("Close err = %v, want nil", err)
	}
}

func TestErrorFormatsWithPrefix(t *testing.T) {
	e := &pdfium.Error{Message: "boom"}
	if got, want := e.Error(), "pdfium: boom"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}
