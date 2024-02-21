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

const apptitle = "BOXES4 version 0.5"
const developmentversion = false

const copyrite = "Copyright © 2024 Bob Stammers"

var cfgfile = flag.String("cfg", "", "Path to YAML configuration file")
var cssfile = flag.String("css", "", "Path to extra CSS file")
var serveport = flag.String("port", "", "HTTP port to serve on")
var dbx = flag.String("db", "", "Path to database file")
var silent = flag.Bool("silent", false, "Suppress terminal output")
var dumpdb = flag.String("export", "", "Dump database to this file")
var loaddb = flag.String("import", "", "Load database from this file")
var makedb = flag.String("makedb", "", "Make a database in this new file")

// Be sure to set these correctly for production releases!
var debug = flag.Bool("debug", developmentversion, "Show debug messages")
var nolocal = flag.Bool("nolocal", developmentversion, "Suppress autostart of browser window")

var DBH *sql.DB
var runvars AppVars
var prefs userpreferences

func main() {

	if !*silent {
		fmt.Printf("%v - %v\n", apptitle, copyrite)
	}
	flag.Parse()

	if *makedb != "" {
		makedatabase(*makedb)
		fmt.Printf("New database %v created\n", *makedb)
		return
	}

	setConfigName()
	loadConfiguration(cfgfile)
	setCssName()
	loadCSS(cssfile)

	initTemplates()

	setDatabaseName()

	var err error
	if _, err = os.Stat(*dbx); err != nil {
		fmt.Printf("Can't access database %v - %v\n", *dbx, err)
		return
	}

	DBH, err = sql.Open("sqlite3", *dbx)
	checkerr(err)

	if *loaddb != "" {
		loaddatabase(*loaddb)
		fmt.Printf("Database %v loaded from %v\n", *dbx, *loaddb)
		return
	}
	if *dumpdb != "" {
		dumpdatabase(*dumpdb)
		fmt.Printf("Database %v dumped to %v\n", *dbx, *dumpdb)
		return
	}

	checkDatabaseVersion(*dbx)
	adbx, err := filepath.Abs(*dbx)
	checkerr(err)
	if !*silent {
		fmt.Printf("Database is %v\n", adbx)
	}
	if *serveport != "" {
		prefs.HttpPort = *serveport
	} else if prefs.HttpPort == "" {
		prefs.HttpPort = "8081"
	}

	if !*silent {
		fmt.Println("Serving on port " + prefs.HttpPort)
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
	http.HandleFunc("/theme", ajax_setTheme)
	http.HandleFunc("/ps", ajax_setPagesize)

	if !*nolocal {
		browser.OpenURL("http://127.0.0.1:" + prefs.HttpPort)
	}
	log.Fatal(http.ListenAndServe(":"+prefs.HttpPort, nil))

}
