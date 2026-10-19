//go:build js

package main

import (
	"database/sql"
	"fmt"
	"strings"
	"syscall/js"

	_ "nimbleterminal.dev/ntshim/duckdb"
)

func main() {
	out := func(s string) {
		div := js.Global().Get("document").Call("getElementById", "out")
		div.Set("innerHTML", div.Get("innerHTML").String()+s+"<br/>")
	}

	out("opening :memory: connection…")
	db, err := sql.Open("duckdb", ":memory:")
	if err != nil {
		out("ERROR open: " + err.Error())
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		out("ERROR ping: " + err.Error())
		return
	}
	out("ping ok")

	mustExec(db, out, `CREATE TABLE fruits (name VARCHAR, qty INTEGER, when_picked TIMESTAMP)`)
	mustExec(db, out, `INSERT INTO fruits VALUES ('apple', 3, TIMESTAMP '2026-05-01 09:30:00'),
	                                              ('banana', 5, TIMESTAMP '2026-05-02 10:00:00'),
	                                              ('cherry', 8, TIMESTAMP '2026-05-03 11:15:00')`)

	out("running SELECT * FROM fruits …")
	rows, err := db.Query(`SELECT name, qty, when_picked FROM fruits WHERE qty >= ? ORDER BY name`, 4)
	if err != nil {
		out("ERROR query: " + err.Error())
		return
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var name string
		var qty int64
		var picked sql.NullTime
		if err := rows.Scan(&name, &qty, &picked); err != nil {
			out("ERROR scan: " + err.Error())
			return
		}
		lines = append(lines, fmt.Sprintf("%-8s %3d  %s", name, qty, picked.Time.Format("2006-01-02 15:04:05")))
	}
	if err := rows.Err(); err != nil {
		out("ERROR rows.Err: " + err.Error())
		return
	}
	out("<pre>" + strings.Join(lines, "\n") + "</pre>")
	out(fmt.Sprintf("done — %d rows", len(lines)))

	// Keep the program alive so the page doesn't show "wasm finished".
	select {}
}

func mustExec(db *sql.DB, out func(string), q string) {
	if _, err := db.Exec(q); err != nil {
		out("ERROR exec: " + err.Error() + " (" + q + ")")
		panic(err)
	}
}
