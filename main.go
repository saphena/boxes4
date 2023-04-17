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

	n, _ := strconv.Atoi(r.FormValue("PAGESIZE"))
	if r.FormValue("PAGESIZE") != "" {
		return n
	}
	return 20
}

func rangeoffset(r *http.Request) int {

	n, _ := strconv.Atoi(r.FormValue("OFFSET"))
	return n

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

func emit_page_anchors(w http.ResponseWriter, r *http.Request, cmd string, totrows, offset, pagesize int) {

	if pagesize < 1 {
		return
	}
	numPages := totrows / pagesize
	if numPages*pagesize < totrows {
		numPages++
	}
	if numPages <= 1 {
		return
	}
	varx := ""
	vars := []string{"FIND", "ORDER", "DESC", "PAGESIZE"}
	for _, v := range vars {
		if r.FormValue(v) != "" {
			if varx != "" {
				varx += "&"
			}
			varx += v + "=" + r.FormValue(v)
		}
	}

	fmt.Fprintf(w, `<div class="pagelinks">`)
	thisPage := (offset / pagesize) + 1
	if thisPage > 1 {
		prevPageOffset := (thisPage * pagesize) - (2 * pagesize)
		fmt.Fprintf(w, `&nbsp;&nbsp;<a id="prevpage" href="/%v?%v&OFFSET=%v" title="Previous page">%v</a>&nbsp;&nbsp;`, cmd, varx, prevPageOffset, ArrowPrevPage)
	}
	minPage := 1
	if thisPage > MaxAdjacentPagelinks {
		minPage = thisPage - MaxAdjacentPagelinks
	}
	maxPage := numPages
	if thisPage < numPages-MaxAdjacentPagelinks {
		maxPage = thisPage + MaxAdjacentPagelinks
	}
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		if pageNum == 1 || pageNum == numPages || (pageNum >= minPage && pageNum <= maxPage) {
			if pageNum == thisPage {
				fmt.Fprintf(w, "[ <strong>%v</strong> ]", thisPage)
			} else {
				pOffset := (pageNum * pagesize) - pagesize

				fmt.Fprintf(w, `[<a href="/%v?%v&OFFSET=%v" title="">%v</a>]`, cmd, varx, pOffset, strconv.Itoa(pageNum))
			}
		} else if pageNum == thisPage-(MaxAdjacentPagelinks+1) || pageNum == thisPage+MaxAdjacentPagelinks+1 {
			fmt.Fprintf(w, " ... ")
		}
	}
	if thisPage < numPages {
		nextPageOffset := (thisPage * pagesize)
		fmt.Fprintf(w, `&nbsp;&nbsp;<a id="nextpage" href="/%v?%v&OFFSET=%v" title="Next page">%v</a>&nbsp;&nbsp;`, cmd, varx, nextPageOffset, ArrowNextPage)
	}

	fmt.Fprint(w, `<select onchange="changepagesize(this);">`)
	pagesizes := []int{0, 20, 40, 60, 100}
	for _, ps := range pagesizes {
		fmt.Fprintf(w, `<option value="%v" `, ps)
		if ps == pagesize {
			fmt.Fprint(w, " selected ")
		}
		fmt.Fprint(w, `>`)
		if ps < 1 {
			fmt.Fprint(w, "show all")
		} else {
			fmt.Fprintf(w, "pagesize %v", ps)
		}
		fmt.Fprint(w, "</option>")
	}
	fmt.Fprint(w, "</select>")

	fmt.Fprintf(w, `</div>`)
}

func showpartners(w http.ResponseWriter, r *http.Request) {

	start_html(w)

	sqlx := "SELECT DISTINCT TRIM(owner), COUNT(TRIM(owner)) AS numdocs FROM contents "
	sqlx += "GROUP BY TRIM(owner) "
	if r.FormValue("PTNR") != "" {
		sqlx += "HAVING TRIM(owner) = '" + r.FormValue("PTNR") + "' "
	}

	if r.FormValue("ORDER") != "" {
		sqlx += "ORDER BY " + r.FormValue("ORDER")
		if r.FormValue("DESC") != "" {
			sqlx += " DESC"
		}
	}

	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	nrex := 0
	for rows.Next() {
		nrex++
	}
	rows.Close()

	ps := pagesize(r)
	ros := 0
	if ps > 0 {
		sqlx += " LIMIT "
		ros = rangeoffset(r)
		if ros > 0 {
			sqlx += strconv.Itoa(ros) + ", "
		}
		sqlx += strconv.Itoa(ps)
	}

	rows, err = DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	emit_page_anchors(w, r, "partners", nrex, ros, ps)

	var plv partnerlistvars

	html, err := template.New("").Parse(partnerlisthdr)
	if err != nil {
		panic(err)
	}
	plv.Desc = r.FormValue("DESC") != r.FormValue("ORDER")
	plv.NumOrder = r.FormValue("ORDER") == "numdocs"
	html.Execute(w, plv)

	html, err = template.New("").Parse(partnerlistline)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		rows.Scan(&plv.Partner, &plv.NumFiles)
		err := html.Execute(w, plv)
		if err != nil {
			panic(err)
		}
	}
	fmt.Fprint(w, partnerlisttrailer)
}

func find(w http.ResponseWriter, r *http.Request) {

	start_html(w)

	var sqlx = ` FROM contents LEFT JOIN boxes ON contents.boxid=boxes.boxid `
	if r.FormValue("FIND") != "" {
		sqlx += `WHERE ((contents.boxid = '?')
		OR (boxes.storeref = '?') 
		OR (boxes.overview LIKE '%?%')
        OR (contents.owner = '?') 
        OR (contents.client = '?') 
        OR (contents.contents LIKE '%?%') 
        OR (contents.name LIKE '%?%')) 

		`
	}
	sqlx = strings.ReplaceAll(sqlx, "?", strings.ReplaceAll(r.FormValue("FIND"), "'", "''"))
	if r.FormValue("ORDER") != "" {
		sqlx += " ORDER BY TRIM(contents." + r.FormValue("ORDER") + ")"
		if r.FormValue("DESC") != "" {
			sqlx += " DESC"
		}
	}

	// fmt.Println(sqlx)

	FoundRecCount, _ := strconv.Atoi(getValueFromDB("SELECT Count(*) AS Rexx"+sqlx, "Rexx", "0"))

	ps := pagesize(r)
	ros := 0
	if ps > 0 {
		sqlx += " LIMIT "
		ros = rangeoffset(r)
		if ros > 0 {
			sqlx += strconv.Itoa(ros) + ", "
		}
		sqlx += strconv.Itoa(ps)
	}

	flds := "contents.BoxID,contents.Owner,contents.Client,contents.Name,contents.Contents,contents.Review_Date"

	rows, err := DBH.Query("SELECT " + flds + sqlx)
	if err != nil {
		fmt.Printf("Omg! %v\n", sqlx)
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
	res.Found = strconv.Itoa(FoundRecCount)
	html.Execute(w, res)
	emit_page_anchors(w, r, "find", FoundRecCount, ros, ps)

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

	sqlboxid := strings.ReplaceAll(r.FormValue("BOXID"), "'", "''")
	sqlx := "SELECT * FROM boxes WHERE boxid='" + sqlboxid + "'"
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
	showBoxfiles(w, sqlboxid)

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

	fmt.Fprintln(w, "</body></html>")
}

func showBoxfiles(w http.ResponseWriter, boxid string) {

	sqlx := "SELECT owner,client,name,contents,review_date FROM contents WHERE boxid='" + boxid + "'"
	sqlx += " ORDER BY owner,client"
	rows, _ := DBH.Query(sqlx)
	defer rows.Close()

	html, err := template.New("").Parse(boxfileshdr)
	if err != nil {
		panic(err)
	}
	err = html.Execute(w, "")
	if err != nil {
		panic(err)
	}

	var bfv boxfilevars

	nrows := 0

	html, err = template.New("").Parse(boxfilesline)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		rows.Scan(&bfv.Partner, &bfv.Client, &bfv.Name, &bfv.Contents, &bfv.Date)
		err = html.Execute(w, bfv)
		if err != nil {
			panic(err)
		}

		nrows++
	}
	html, err = template.New("").Parse(boxfilestrailer)
	html.Execute(w, "")
	if err != nil {
		panic(err)
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
	http.HandleFunc("/partners", showpartners)

	log.Fatal(http.ListenAndServe(":"+*serveport, nil))

}
