// smoke.mjs — headless approximation of the browser check for the
// flintchart-compile example. It cannot exercise the Go/WASM side (that
// needs a real browser — see README.md's "Manual verification" section),
// but it does load the vendored bundle exactly as flintchart-shim.js runs
// in a page and drives it with the same sample input as main.go, proving
// the bundle itself compiles and returns a well-formed envelope.
//
// Usage: node smoke.mjs   (run from this directory, after `task build`
// or `go tool ntshim-assets . --shim=flintchart` has populated
// ntshim/flintchart/flintchart-shim.js)

const sampleInput = JSON.stringify({
  data: {
    values: [
      { month: "Jan", revenue: 120 },
      { month: "Feb", revenue: 180 },
      { month: "Mar", revenue: 95 },
      { month: "Apr", revenue: 210 },
    ],
  },
  semantic_types: { revenue: { semanticType: "Price", unit: "USD" } },
  chart_spec: {
    chartType: "Bar Chart",
    encodings: { x: { field: "month" }, y: { field: "revenue" } },
    baseSize: { width: 60, height: 16 },
  },
});

await import("./ntshim/flintchart/flintchart-shim.js");
const ns = globalThis.ntshim.flintchart;

await ns.ready;

const raw = ns.compile(sampleInput);
const envelope = JSON.parse(raw);

if (envelope.error) {
  console.error("compile error:", envelope.error.message);
  process.exit(1);
}
if (!envelope.size || typeof envelope.size.width !== "number") {
  console.error("malformed envelope, missing size:", raw);
  process.exit(1);
}

console.log("version:", ns.version);
console.log("warnings:", (envelope.warnings ?? []).length);
console.log("size:", envelope.size);
