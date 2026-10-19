// Command ntshim-assets copies the JS-side bridge files for one or
// more ntshim subpackages into a target site directory.
//
// Usage:
//
//	ntshim-assets <web-dir> --shim=duckdb [--shim=…] [--cdn|--vendored]
//
// Outputs land under <web-dir>/ntshim/<shim>/.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
)

func main() {
	var (
		shims    []string
		cdn      bool
		vendored bool
	)
	flag.StringSliceVar(&shims, "shim", nil, "Shim to install (repeat for multiple). Required.")
	flag.BoolVar(&cdn, "cdn", true, "Load DuckDB-Wasm + apache-arrow from JSDelivr (default)")
	flag.BoolVar(&vendored, "vendored", false, "Copy pinned bundles into <web-dir>/ntshim/<shim>/")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: ntshim-assets <web-dir> --shim=NAME [--shim=NAME ...]")
		os.Exit(2)
	}
	if len(shims) == 0 {
		fmt.Fprintln(os.Stderr, "error: at least one --shim is required")
		os.Exit(2)
	}
	if vendored {
		cdn = false
	}
	_ = cdn // reserved for future use; flag still surfaces in --help

	webDir := flag.Arg(0)
	for _, shim := range shims {
		if err := installShim(shim, webDir, vendored); err != nil {
			fmt.Fprintf(os.Stderr, "installing %s: %v\n", shim, err)
			os.Exit(1)
		}
	}
}

func installShim(shim, webDir string, vendored bool) error {
	// AudioBuffer scheduling is implemented entirely in Go/syscall/js.
	if shim == "audio" {
		return nil
	}
	srcRoot := filepath.Join("assets", shim)
	if _, err := fs.Stat(shimAssets, srcRoot); err != nil {
		return fmt.Errorf("unknown shim %q (no embedded assets under %s)", shim, srcRoot)
	}
	dstRoot := filepath.Join(webDir, "ntshim", shim)
	if err := os.MkdirAll(dstRoot, 0o755); err != nil {
		return err
	}
	return fs.WalkDir(shimAssets, srcRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(srcRoot, p)
		if d.IsDir() {
			if rel == "." {
				return nil
			}
			// In --cdn mode, skip creating the vendored/ subdirectory.
			if !vendored && rel == "vendored" {
				return fs.SkipDir
			}
			return os.MkdirAll(filepath.Join(dstRoot, rel), 0o755)
		}
		// In --cdn mode, skip files under a "vendored/" subdirectory.
		if !vendored && strings.HasPrefix(rel, "vendored"+string(filepath.Separator)) {
			return nil
		}
		// Bridge-file swap:
		//   --cdn:      write duckdb-shim.js as-is; skip duckdb-shim-vendored.js.
		//   --vendored: skip duckdb-shim.js (CDN bridge); write duckdb-shim-vendored.js
		//               AS duckdb-shim.js so consumer HTML needs no changes.
		if rel == "duckdb-shim.js" && vendored {
			return nil // skip CDN bridge in vendored mode
		}
		if rel == "duckdb-shim-vendored.js" {
			if !vendored {
				return nil // skip vendored bridge in CDN mode
			}
			// Write vendored bridge AS duckdb-shim.js so consumer HTML
			// doesn't need to change its <script> src.
			data, err := fs.ReadFile(shimAssets, p)
			if err != nil {
				return err
			}
			out := filepath.Join(dstRoot, "duckdb-shim.js")
			fmt.Printf("write %s (vendored)\n", out)
			return os.WriteFile(out, data, 0o644)
		}
		data, err := fs.ReadFile(shimAssets, p)
		if err != nil {
			return err
		}
		out := filepath.Join(dstRoot, rel)
		fmt.Printf("write %s\n", out)
		return os.WriteFile(out, data, 0o644)
	})
}
