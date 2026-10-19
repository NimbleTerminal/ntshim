//go:build !js

package flintchart_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"nimbleterminal.dev/ntshim/flintchart"
)

func TestNonJSStubs(t *testing.T) {
	if _, err := flintchart.Compile(context.Background(), []byte(`{}`)); !errors.Is(err, flintchart.ErrNonJSPlatform) {
		t.Fatalf("Compile stub: %v", err)
	}
	if _, err := flintchart.Version(); !errors.Is(err, flintchart.ErrNonJSPlatform) {
		t.Fatalf("Version stub: %v", err)
	}
}

func TestErrorPrefix(t *testing.T) {
	err := &flintchart.Error{Message: "boom"}
	if !strings.HasPrefix(err.Error(), "flintchart: ") {
		t.Fatalf("prefix missing: %v", err)
	}
}
