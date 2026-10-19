//go:build js

package main

import (
	"context"
	"fmt"
	"strings"
	"syscall/js"

	"nimbleterminal.dev/ntshim/duckdb"
	"github.com/apache/arrow-go/v18/arrow/array"
)

func main() {
	out := func(s string) {
		div := js.Global().Get("document").Call("getElementById", "out")
		div.Set("innerHTML", div.Get("innerHTML").String()+s+"<br/>")
	}

	ctx := context.Background()
	cli, err := duckdb.Open(ctx, ":memory:")
	if err != nil {
		out("ERROR open: " + err.Error())
		return
	}
	defer cli.Close()

	for _, q := range []string{
		`CREATE TABLE nums (n INTEGER, sq INTEGER)`,
		`INSERT INTO nums SELECT i, i*i FROM range(1, 6) t(i)`,
	} {
		if _, err := cli.Exec(ctx, q); err != nil {
			out("ERROR exec: " + err.Error())
			return
		}
	}

	rr, err := cli.QueryArrow(ctx, `SELECT n, sq FROM nums ORDER BY n`)
	if err != nil {
		out("ERROR query: " + err.Error())
		return
	}
	defer rr.Release()

	var lines []string
	for rr.Next() {
		rec := rr.Record()
		n := rec.Column(0).(*array.Int32)
		sq := rec.Column(1).(*array.Int32)
		for i := 0; i < int(rec.NumRows()); i++ {
			lines = append(lines, fmt.Sprintf("n=%d sq=%d", n.Value(i), sq.Value(i)))
		}
	}
	if err := rr.Err(); err != nil {
		out("ERROR rr.Err: " + err.Error())
		return
	}
	out("<pre>" + strings.Join(lines, "\n") + "</pre>")
	select {}
}
