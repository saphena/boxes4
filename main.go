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
var runvars AppVars
var prefs userpreferences

func main() {

	cfgfile := flag.String("cfg", "", "Path to YAML configuration file")
	serveport := flag.String("port", "", "HTTP port to serve on")
	flag.Parse()
	loadConfiguration(cfgfile)

	if *serveport != "" {
		prefs.HttpPort = *serveport
	} else if prefs.HttpPort == "" {
		prefs.HttpPort = "8081"
	}

	initTemplates()

	var err error
	DBH, err = sql.Open("sqlite3", dbx)
	checkerr(err)
	http.HandleFunc("/", show_search)

	http.HandleFunc("/search", show_search)
	http.HandleFunc("/find", exec_search)
	http.HandleFunc("/about", about)
	http.HandleFunc("/boxes", showboxes)
	http.HandleFunc("/check", check_database)
	http.HandleFunc("/csvexp", csvexp)
	http.HandleFunc("/jsonexp", jsonexp)
	http.HandleFunc("/owners", showowners)
	http.HandleFunc("/params", show_search_params)
	http.HandleFunc("/locations", showlocations)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/update", update)
	http.HandleFunc("/users", showusers)
	http.HandleFunc("/userx", ajax_users)
	http.HandleFunc("/secret", secret)

	log.Fatal(http.ListenAndServe(":"+prefs.HttpPort, nil))

}
