package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The embedded files are what consumers actually install. Check their public
// namespace and types as well as their agreement with the source bridges.
func TestInstalledBridges(t *testing.T) {
	for _, shim := range []string{"duckdb", "pdfium", "flintchart"} {
		t.Run(shim, func(t *testing.T) {
			dir := t.TempDir()
			if err := installShim(shim, dir, false); err != nil {
				t.Fatal(err)
			}
			names := []string{shim + "-shim.js", shim + "-shim.d.ts"}
			if shim == "duckdb" {
				names = append(names, "duckdb-bridge.js", "haybarn-shim.js")
				names = append(names, "haybarn/duckdb-browser.mjs", "haybarn/PROVENANCE.txt", "haybarn/LICENSE")
			}
			for _, name := range names {
				installed, err := os.ReadFile(filepath.Join(dir, "ntshim", shim, name))
				if err != nil {
					t.Fatal(err)
				}
				source, err := os.ReadFile(filepath.Join("..", "..", "web", shim, name))
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(installed, source) {
					t.Fatalf("installed %s differs from web source", name)
				}
				if !strings.Contains(name, "/") && (!bytes.Contains(installed, []byte("ntshim")) || bytes.Contains(installed, []byte("boobaShim"))) {
					t.Fatalf("installed %s has the wrong browser namespace", name)
				}
			}
			for _, absent := range []string{"booba-shim", filepath.Join("ntshim", shim, "vendored"), filepath.Join("ntshim", shim, "duckdb-shim-vendored.js")} {
				if _, err := os.Stat(filepath.Join(dir, absent)); !os.IsNotExist(err) {
					t.Fatalf("unexpected output %s: %v", absent, err)
				}
			}
		})
	}
}

func TestVendoredDuckDB(t *testing.T) {
	dir := t.TempDir()
	if err := installShim("duckdb", dir, true); err != nil {
		t.Fatal(err)
	}
	bridge, err := os.ReadFile(filepath.Join(dir, "ntshim", "duckdb", "duckdb-shim.js"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join("..", "..", "web", "duckdb", "duckdb-shim-vendored.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bridge, want) {
		t.Fatal("vendored mode did not install the vendored bridge under the standard filename")
	}
	for _, absent := range []string{"haybarn-shim.js", "haybarn"} {
		if _, err := os.Stat(filepath.Join(dir, "ntshim", "duckdb", absent)); !os.IsNotExist(err) {
			t.Fatalf("CDN-only Haybarn file installed in vendored mode: %s, %v", absent, err)
		}
	}
	for _, name := range []string{"duckdb-mvp.wasm", "duckdb-eh.wasm", "duckdb-browser.mjs", "apache-arrow.mjs"} {
		info, err := os.Stat(filepath.Join(dir, "ntshim", "duckdb", "vendored", name))
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() == 0 {
			t.Fatalf("empty vendored asset %s", name)
		}
	}
}

func TestHaybarnBrowserModuleProvenance(t *testing.T) {
	bundle, err := shimAssets.ReadFile("assets/duckdb/haybarn/duckdb-browser.mjs")
	if err != nil {
		t.Fatal(err)
	}
	provenance, err := shimAssets.ReadFile("assets/duckdb/haybarn/PROVENANCE.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("\nsha256: %x\n", sha256.Sum256(bundle))
	if !bytes.Contains(provenance, []byte(want)) {
		t.Fatal("Haybarn browser module differs from its recorded checksum")
	}
	if bytes.Contains(bundle, []byte(`from"apache-arrow"`)) {
		t.Fatal("Haybarn browser module still has bare Arrow imports")
	}
}

func TestAudioNeedsNoAssets(t *testing.T) {
	dir := t.TempDir()
	if err := installShim("audio", dir, false); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("audio created assets: %v, %v", entries, err)
	}
}

func TestUnknownShim(t *testing.T) {
	if err := installShim("missing", t.TempDir(), false); err == nil {
		t.Fatal("unknown shim succeeded")
	}
}
