package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const dbx = "boxes.db"

var DBH *sql.DB
var err error
var runvars AppVars

func getValueFromDB(sqlx string, col string, defval string) string {

	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var res string
	if !rows.Next() {
		return defval
	}
	rows.Scan(&res)
	return res
}

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
		tsfmt := "2006-01-02T15:04:05Z"
		ts, err := time.Parse(tsfmt, lastUpdated)
		if err != nil {
			fmt.Fprint(w, err)
		}
		fmt.Fprintf(w, "Database last updated <strong>%v</strong>", ts.Format("Monday 2 Jan 2006 @ 3:04pm"))
	} else {
		fmt.Fprint(w, "Not updated")
	}
	tables := []string{"BOXES", "CONTENTS", "HISTORY", "LOCATIONS", "USERS"}

	fmt.Fprint(w, "<ul>")
	for _, tab := range tables {
		sqlx := "SELECT Count(*) As Rex FROM " + tab
		rex := getValueFromDB(sqlx, "Rex", "0")
		fmt.Fprintf(w, "<li>Table %v has <strong>%v</strong> records</li>", tab, rex)
	}
	fmt.Fprint(w, "</ul>")

}

func pagesize(r *http.Request) int {

	return 60
}

func rangeoffset(r *http.Request) int {

	return 0

}

func start_html(w http.ResponseWriter) {

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

}

func order_dir(r *http.Request, field string) string {
	if (r.FormValue("ORDER") != field) || (r.FormValue("ORDER") == r.FormValue("DESC")) {
		return string("")
	}
	return "&amp;DESC=" + r.FormValue("ORDER")
}

func find(w http.ResponseWriter, r *http.Request) {

	start_html(w)

	var sqlx = ` FROM contents LEFT JOIN boxes ON contents.boxid=boxes.boxid `
	if r.FormValue("FIND") != "" {
		sqlx += `WHERE ((contents.boxid = '?') 
        OR (contents.owner = '?') 
        OR (contents.client = '?') 
        OR (contents.contents LIKE '%?%') 
        OR (contents.name LIKE '%?%')) 

		`
	}
	sqlx = strings.ReplaceAll(sqlx, "?", strings.ReplaceAll(r.FormValue("FIND"), "'", "''"))
	if r.FormValue("ORDER") != "" {
		sqlx += " ORDER BY contents." + r.FormValue("ORDER")
		if r.FormValue("DESC") != "" {
			sqlx += " DESC"
		}
	}

	ps := pagesize(r)
	if ps > 0 {
		sqlx += " LIMIT "
		os := rangeoffset(r)
		if os > 0 {
			sqlx += strconv.Itoa(os) + " "
		}
		sqlx += strconv.Itoa(ps)
	}

	// fmt.Println(sqlx)

	FoundRecCount := getValueFromDB("SELECT Count(*) AS Rexx"+sqlx, "Rexx", "0")

	flds := "contents.BoxID,contents.Owner,contents.Client,contents.Name,contents.Contents,contents.Review_Date"

	rows, err := DBH.Query("SELECT " + flds + sqlx)
	if err != nil {
		panic(err)
	}
	html, err := template.New("main").Parse(searchResultsHdr)
	if err != nil {
		panic(err)
	}
	var res searchResultsVar
	res.Boxid = order_dir(r, "boxid")
	res.Partner = order_dir(r, "owner")
	res.Client = order_dir(r, "client")
	res.Name = order_dir(r, "name")
	res.Date = order_dir(r, "review_date")
	res.Find = r.FormValue("FIND")
	res.Found = FoundRecCount
	html.Execute(w, res)

	for rows.Next() {
		rows.Scan(&res.Boxid, &res.Partner, &res.Client, &res.Name, &res.Contents, &res.Date)
		html, _ = template.New("main").Parse(searchResultsLine)
		html.Execute(w, res)
	}
	html, _ = template.New("main").Parse(searchResultsTrailer)
	html.Execute(w, "")

}

func showbox(w http.ResponseWriter, r *http.Request) {

	if r.FormValue("BOXID") == "" {
		search(w, r)
		return
	}

	start_html(w)

	sqlx := "SELECT * FROM boxes WHERE boxid='" + strings.ReplaceAll(r.FormValue("BOXID"), "'", "''") + "'"
	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var bv boxvars
	if !rows.Next() {
		fmt.Fprintf(w, "<p>Bugger! %v</p>", r.FormValue("BOXID"))
		return
	}
	var mindate, maxdate string
	rows.Scan(&bv.Storeref, &bv.Boxid, &bv.Location, &bv.Contents, &bv.NumFiles, &mindate, &maxdate)
	bv.Date = mindate + " to " + maxdate
	html, err := template.New("main").Parse(boxhtml)
	if err != nil {
		panic(err)
	}
	err = html.Execute(w, bv)
	if err != nil {
		panic(err)
	}

}
func search(w http.ResponseWriter, r *http.Request) {

	start_html(w)

	searchVars.Apptitle = "DOCUMENT ARCHIVES"
	searchVars.NumBoxes, _ = strconv.Atoi(getValueFromDB("SELECT Count(*) As Rex FROM boxes", "Rex", "-1"))
	searchVars.NumDocs, _ = strconv.Atoi(getValueFromDB("SELECT Count(*) As Rex FROM contents", "Rex", "-1"))
	searchVars.NumLocns, _ = strconv.Atoi(getValueFromDB("SELECT Count(*) As Rex FROM locations", "Rex", "-1"))

	html, err := template.New("main").Parse(searchHTML)
	if err != nil {
		panic(err)
	}

	html.Execute(w, searchVars)

	listBoxes(w)

	fmt.Fprintln(w, "</body></html>")
}

func listBoxes(w http.ResponseWriter) {
	sqlx := "SELECT storeref,boxid FROM boxes"

	rows, _ := DBH.Query(sqlx)
	defer rows.Close()
	var storeref, boxid string

	nrows := 0
	for rows.Next() {
		rows.Scan(&storeref, &boxid)
		fmt.Fprintf(w, "Found box %v<br>\n", boxid)
		nrows++
		if nrows >= 10 {
			break
		}
	}

}
func main() {

	serveport := flag.String("port", "", "HTTP port to serve on")
	flag.Parse()
	if *serveport == "" {
		*serveport = "8081"
	}
	runvars = AppVars{"DOCUMENT ARCHIVES", basicMenu}
	DBH, err = sql.Open("sqlite3", dbx)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", search)

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})

	http.HandleFunc("/search", search)
	http.HandleFunc("/find", find)
	http.HandleFunc("/about", about)
	http.HandleFunc("/showbox", showbox)

	log.Fatal(http.ListenAndServe(":"+*serveport, nil))

}
