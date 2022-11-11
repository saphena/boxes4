package main

import (
	"database/sql"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

const dbx = "boxes.db"

var DBH *sql.DB
var err error
var runvars AppVars

func search(w http.ResponseWriter, r *http.Request) {
	sqlx := "SELECT storeref,boxid FROM boxes"

	rows, _ := DBH.Query(sqlx)
	defer rows.Close()
	var storeref, boxid string

	var ht string
	if true {
		ht = html1 + css + html2 + basicMenu + "</div>"
	} else {
		ht = html1 + css + html2 + updateMenu + "</div>"
	}
	html, err := template.New("main").Parse(ht)
	if err != nil {
		panic(err)
	}

	html.Execute(w, runvars)
	for rows.Next() {
		rows.Scan(&storeref, &boxid)
		fmt.Fprintf(w, "Found box %v<br>\n", boxid)
	}
	fmt.Fprintln(w, "</body></html>")
}
func main() {

	runvars = AppVars{"DOCUMENT ARCHIVES", basicMenu}
	DBH, err = sql.Open("sqlite3", dbx)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})

	http.HandleFunc("/search", search)

	log.Fatal(http.ListenAndServe(":8081", nil))

}
