//go:build js

package duckdb

import (
	"bytes"
	"context"
	"errors"
	"syscall/js"

	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/ipc"
)

// Client is a native (non-database/sql) handle to a DuckDB-Wasm connection.
// It is safe to use from one goroutine at a time; concurrent calls share
// the underlying JS connection and must be serialized by the caller.
type Client struct {
	handle js.Value
	open   bool
}

// Open opens a connection. The DSN format matches the database/sql driver:
//
//	<path>?key=val&key=val   or   ":memory:"
//
// Files referenced by <path> must be pre-registered via the JS bridge
// (window.ntshim.duckdb.registerFileURL or registerFileBuffer).
func Open(ctx context.Context, dsn string) (*Client, error) {
	_ = ctx // ctx is reserved for cancellation in a future release.
	res, err := awaitPromise(bridge().Call("open", dsn))
	if err != nil {
		return nil, err
	}
	return &Client{handle: res, open: true}, nil
}

// Close releases the JS-side connection. The Go side is marked closed
// only after the JS close promise resolves: the bridge removes its
// handle entry after conn.close() resolves, so a rejected close leaves
// the connection open on both sides and the caller can retry.
func (c *Client) Close() error {
	if !c.open {
		return nil
	}
	if _, err := awaitPromise(bridge().Call("close", c.handle)); err != nil {
		return err
	}
	c.open = false
	return nil
}

// QueryArrow runs sql with positional params and returns an ipc.Reader
// over the Arrow IPC stream produced by DuckDB-Wasm. The caller MUST call
// Release() on the reader when done.
func (c *Client) QueryArrow(ctx context.Context, sql string, params ...any) (*ipc.Reader, error) {
	_ = ctx
	if !c.open {
		return nil, errors.New("duckdb: connection closed")
	}
	jsArgs, err := marshalAnySlice(params)
	if err != nil {
		return nil, err
	}
	res, err := awaitPromise(bridge().Call("query", c.handle, sql, jsArgs))
	if err != nil {
		return nil, err
	}
	n := res.Get("length").Int()
	buf := make([]byte, n)
	js.CopyBytesToGo(buf, res)
	return ipc.NewReader(bytes.NewReader(buf))
}

// Exec runs DDL/DML statements; returns rowsAffected (always 0 — see
// the spec for the DuckDB-Wasm limitation).
func (c *Client) Exec(ctx context.Context, sql string, params ...any) (int64, error) {
	_ = ctx
	if !c.open {
		return 0, errors.New("duckdb: connection closed")
	}
	jsArgs, err := marshalAnySlice(params)
	if err != nil {
		return 0, err
	}
	res, err := awaitPromise(bridge().Call("exec", c.handle, sql, jsArgs))
	if err != nil {
		return 0, err
	}
	return int64(res.Get("rowsAffected").Int()), nil
}

// marshalAnySlice is the native-API counterpart to marshalArgs: same
// type coverage, different input shape.
func marshalAnySlice(params []any) (js.Value, error) {
	arr := js.Global().Get("Array").New(len(params))
	for i, p := range params {
		v, err := marshalValue(p)
		if err != nil {
			return js.Value{}, err
		}
		arr.SetIndex(i, v)
	}
	return arr, nil
}

// Keep the array import alive for downstream consumers of QueryArrow that
// will reach for arrow.Record helpers.
var _ = array.NewRecord
