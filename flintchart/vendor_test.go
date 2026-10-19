//go:build !js

package flintchart_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

// pair is one vendored copy of the shim bundle plus its sibling PROVENANCE
// file, both relative to this package's directory.
type pair struct {
	bundle     string
	provenance string
}

var vendoredPairs = []pair{
	{bundle: "../web/flintchart/flintchart-shim.js", provenance: "../web/flintchart/PROVENANCE.txt"},
	{bundle: "../cmd/ntshim-assets/assets/flintchart/flintchart-shim.js", provenance: "../cmd/ntshim-assets/assets/flintchart/PROVENANCE.txt"},
}

// provenanceSHA256 extracts the value of the "sha256:" line from a
// PROVENANCE.txt file's contents.
func provenanceSHA256(t *testing.T, raw []byte) string {
	t.Helper()
	for _, line := range strings.Split(string(raw), "\n") {
		if rest, ok := strings.CutPrefix(line, "sha256:"); ok {
			return strings.TrimSpace(rest)
		}
	}
	t.Fatalf("no sha256: line found in provenance file")
	return ""
}

// TestVendoredBundlesMatchProvenance guards against the two vendored copies
// of the flint-ntcharts browser shim (web/flintchart and
// cmd/ntshim-assets/assets/flintchart) silently drifting apart, or
// drifting from the sha256 recorded in their own PROVENANCE.txt — either of
// which would mean the shipped bundle no longer matches the reviewed
// flint-ntcharts bundle with the recorded local namespace patch.
func TestVendoredBundlesMatchProvenance(t *testing.T) {
	var (
		bundleShas      [2]string
		provenanceBytes [2][]byte
	)

	for i, p := range vendoredPairs {
		bundle, err := os.ReadFile(p.bundle)
		if err != nil {
			t.Fatalf("reading %s: %v", p.bundle, err)
		}
		sum := sha256.Sum256(bundle)
		bundleShas[i] = hex.EncodeToString(sum[:])

		prov, err := os.ReadFile(p.provenance)
		if err != nil {
			t.Fatalf("reading %s: %v", p.provenance, err)
		}
		provenanceBytes[i] = prov

		want := provenanceSHA256(t, prov)
		if bundleShas[i] != want {
			t.Fatalf("%s: sha256 %s does not match %s (%s)", p.bundle, bundleShas[i], p.provenance, want)
		}
	}

	if bundleShas[0] != bundleShas[1] {
		t.Fatalf("vendored bundles differ: %s (%s) != %s (%s)",
			vendoredPairs[0].bundle, bundleShas[0], vendoredPairs[1].bundle, bundleShas[1])
	}

	if string(provenanceBytes[0]) != string(provenanceBytes[1]) {
		t.Fatalf("PROVENANCE files differ: %s != %s", vendoredPairs[0].provenance, vendoredPairs[1].provenance)
	}
}
