package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

const dbx = "boxes.db"

var DBH *sql.DB
var err error
var runvars AppVars

func main() {

	serveport := flag.String("port", "", "HTTP port to serve on")
	flag.Parse()
	if *serveport == "" {
		*serveport = "8081"
	}
	runvars = AppVars{`DOCUMENT ARCHIVES`, basicMenu, ""}
	DBH, err = sql.Open("sqlite3", dbx)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", show_search)

	http.HandleFunc("/search", show_search)
	http.HandleFunc("/find", exec_search)
	http.HandleFunc("/about", about)
	http.HandleFunc("/boxes", showboxes)
	http.HandleFunc("/check", check_database)
	http.HandleFunc("/csvexp", csvexp)
	http.HandleFunc("/jsonexp", jsonexp)
	http.HandleFunc("/owners", showowners)
	http.HandleFunc("/locations", showlocations)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/update", update)
	http.HandleFunc("/secret", secret)

	log.Fatal(http.ListenAndServe(":"+*serveport, nil))

}
