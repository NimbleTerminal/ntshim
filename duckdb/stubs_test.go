//go:build !js

package duckdb_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"nimbleterminal.dev/ntshim/duckdb"
)

func TestSqlOpenReturnsErrNonJSPlatform(t *testing.T) {
	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		// sql.Open is lazy; the driver error surfaces on first use.
		t.Fatalf("sql.Open returned unexpected error: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); !errors.Is(err, duckdb.ErrNonJSPlatform) {
		t.Fatalf("Ping err = %v, want ErrNonJSPlatform", err)
	}
}

func TestClientOpenReturnsErrNonJSPlatform(t *testing.T) {
	_, err := duckdb.Open(context.Background(), ":memory:")
	if !errors.Is(err, duckdb.ErrNonJSPlatform) {
		t.Fatalf("Open err = %v, want ErrNonJSPlatform", err)
	}
}
