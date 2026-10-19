package main

import (
	"bytes"
	"os"
	"path/filepath"
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
				if !bytes.Contains(installed, []byte("ntshim")) || bytes.Contains(installed, []byte("boobaShim")) {
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
