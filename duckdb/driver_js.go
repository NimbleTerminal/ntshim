//go:build js

package duckdb

import (
	"database/sql"
	"database/sql/driver"
)

func init() {
	sql.Register("duckdb", &jsDriver{})
}

type jsDriver struct{}

// Open implements driver.Driver. The DSN format matches duckdb-go:
//
//	<path>?key=val&key=val
//
// where <path> is either ":memory:" or a virtual filename previously
// registered via ntshim.duckdb.registerFileURL / registerFileBuffer.
func (jsDriver) Open(dsn string) (driver.Conn, error) {
	res, err := awaitPromise(bridge().Call("open", dsn))
	if err != nil {
		return nil, err
	}
	return &jsConn{handle: res, open: true}, nil
}
