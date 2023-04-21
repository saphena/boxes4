package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const dbx = "boxes.db"

var DBH *sql.DB
var err error
var runvars AppVars

func about(w http.ResponseWriter, r *http.Request) {

	start_html(w)
	fmt.Fprint(w, "<h2>BOXES version 4.0</h2>")
	fmt.Fprint(w, "<p class='copyrite'>Copyright &copy; 2023 Bob Stammers &lt;stammers.bob@gmail.com&gt; </p>")

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	fmt.Fprintf(w, "<p>I'm installed in the folder <strong>%v</strong></p>", exPath)
	lastUpdated := getValueFromDB("SELECT recordedat FROM history ORDER BY recordedat DESC LIMIT 0,1", "recordedat", "")
	if lastUpdated != "" {
		updatedBy := getValueFromDB("SELECT userid FROM history ORDER BY recordedat DESC LIMIT 0,1", "userid", "")
		tsfmt := "2006-01-02T15:04:05Z"
		ts, err := time.Parse(tsfmt, lastUpdated)
		if err != nil {
			fmt.Fprint(w, err)
		}
		fmt.Fprintf(w, "<p>Database last updated <strong>%v</strong> by '%v'</p>", ts.Format("Monday 2 Jan 2006 @ 3:04pm"), updatedBy)
	} else {
		fmt.Fprint(w, "Not updated")
	}
	fmt.Fprint(w, `<p>Click [update] above and login as a user with CONTROLLER accesslevel to get more info. `)
	var uids []string
	rows, err := DBH.Query("SELECT userid FROM users WHERE accesslevel >= " + strconv.Itoa(ACCESSLEVEL_UPDATE))
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var uid string
		rows.Scan(&uid)
		uids = append(uids, uid)
	}
	fmt.Fprintf(w, ` The following userids have that accesslevel: <strong>%v</strong></p>`, uids)
	tables := []string{"BOXES", "CONTENTS", "HISTORY", "LOCATIONS", "USERS"}

	fmt.Fprint(w, "<ul>")
	for _, tab := range tables {
		sqlx := "SELECT Count(*) As Rex FROM " + tab
		rex := getValueFromDB(sqlx, "Rex", "0")
		fmt.Fprintf(w, "<li>Table %v has <strong>%v</strong> records</li>", tab, rex)
	}
	fmt.Fprint(w, "</ul>")

}

func main() {

	serveport := flag.String("port", "", "HTTP port to serve on")
	flag.Parse()
	if *serveport == "" {
		*serveport = "8081"
	}
	runvars = AppVars{`DOCUMENT ARCHIVES`, basicMenu}
	DBH, err = sql.Open("sqlite3", dbx)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", show_search)

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})

	http.HandleFunc("/search", show_search)
	http.HandleFunc("/find", exec_search)
	http.HandleFunc("/about", about)
	http.HandleFunc("/boxes", showboxes)
	http.HandleFunc("/owners", showowners)

	log.Fatal(http.ListenAndServe(":"+*serveport, nil))

}
