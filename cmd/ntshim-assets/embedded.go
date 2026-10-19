package main

import "embed"

// shimAssets holds the JS/TS bridge files for every shim. The directory
// layout mirrors the install path under <web-dir>/ntshim/<shim>/.
//
//go:embed all:assets
var shimAssets embed.FS
