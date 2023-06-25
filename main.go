package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/browser"
)

const apptitle = "BOXES4 version 0.2"
const copyrite = "Copyright © 2023 Bob Stammers"

var cfgfile = flag.String("cfg", "", "Path to YAML configuration file")
var serveport = flag.String("port", "", "HTTP port to serve on")
var dbx = flag.String("db", "boxes.db", "Path to database file")
var silent = flag.Bool("silent", false, "Suppress terminal output")

// Be sure to set these correctly for production releases!
var debug = flag.Bool("debug", true, "Show debug messages")
var nolocal = flag.Bool("nolocal", true, "Suppress autostart of browser window")

var DBH *sql.DB
var runvars AppVars
var prefs userpreferences

func main() {

	flag.Parse()
	loadConfiguration(cfgfile)

	if *serveport != "" {
		prefs.HttpPort = *serveport
	} else if prefs.HttpPort == "" {
		prefs.HttpPort = "8081"
	}

	if !*silent {
		fmt.Printf("%v - %v\n", apptitle, copyrite)
		fmt.Println("Serving on port " + prefs.HttpPort)
	}

	initTemplates()

	var err error
	if _, err = os.Stat(*dbx); err != nil {
		fmt.Printf("Can't access database %v - %v\n", *dbx, err)
		return
	}
	DBH, err = sql.Open("sqlite3", *dbx)
	checkerr(err)
	checkDatabaseVersion(*dbx)
	adbx, err := filepath.Abs(*dbx)
	checkerr(err)
	if !*silent {
		fmt.Printf("Database is %v\n", adbx)
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
	http.HandleFunc("/params", show_search_params)
	http.HandleFunc("/locations", showlocations)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/update", update)
	http.HandleFunc("/users", showusers)
	http.HandleFunc("/userx", ajax_users)
	http.HandleFunc("/secret", secret)

	if !*nolocal {
		browser.OpenURL("http://127.0.0.1:" + prefs.HttpPort)
	}
	log.Fatal(http.ListenAndServe(":"+prefs.HttpPort, nil))

}
