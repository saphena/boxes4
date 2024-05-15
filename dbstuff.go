package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// These table definitions are used as templates for
// dumping to CSV files, probably for use in spreadsheets.
type table_boxes struct {
	Storeref        string
	Boxid           string
	Location        string
	Overview        string
	NumDocs         int
	Min_Review_date string
	Max_Review_date string
}

type table_contents struct {
	Id          int
	Boxid       string
	Review_date string
	Contents    string
	Owner       string
	Name        string
	Client      string
}

type table_locations struct {
	Location string
}

type table_history struct {
	Recordedat string
	Userid     string
	TheSQL     string
	TheResult  int
}

type table_owners struct {
	Owner string
	Name  string
}

type table_users struct {
	Userid      string
	Userpass    string
	Accesslevel int
}

var boxes_with_contents int
var boxes_with_contents_updates int

var orphaned_boxes []string
var empty_boxes []string
var big_boxes []string
var childless_owners []string

// buildOwnersTable is called only once to populate the new table owners.
// Repeated calls will do no harm.
func buildOwnersTable() {

	sqlx := "INSERT OR IGNORE INTO owners(owner,name) SELECT DISTINCT TRIM(owner), LOWER(TRIM(owner)) FROM contents ORDER BY TRIM(owner);"
	_, err := DBH.Exec(sqlx)
	checkerr(err)
}

func setDatabaseName() {

	if *dbx != "" { // Manually set on commandline
		return
	}
	exe, err := os.Executable()
	checkerr(err)
	exex := filepath.Ext(exe)
	if exex != "" {
		exe = strings.TrimSuffix(exe, exex)
	}
	*dbx = exe + ".db"
}

// syncOwner is called whenever an owner field is updated in contents in order
// to maintain the separate owners table. No entries are deleted from the table.
func syncOwner(owner string) {

	sqlx := "INSERT OR IGNORE INTO owners(owner,name) VALUES('" + safesql(owner) + "','" + strings.ToLower((safesql(owner))) + "');"
	DBExec(sqlx)

}

func DBEscape(arg string) string {

	return strings.ReplaceAll(arg, "'", "''")
}

func DBExec(sqlx string) sql.Result {

	res, err := DBH.Exec(sqlx)
	if err != nil {
		fmt.Printf("DBExec = %v [%v]\n", sqlx, err)
		//		panic(err)
	} else {
		re := regexp.MustCompile(`(?i)^(ALTER|CREATE|DELETE|DROP|INSERT|REPLACE|UPDATE|UPSERT)`)
		if re.MatchString(sqlx) {
			markDBTouch(sqlx)
		}
	}
	return res

}

func markDBTouch(sqlx string) {

	touch := "INSERT INTO history (recordedat,userid,thesql,theresult) VALUES("
	dt := time.Now().Format(time.RFC3339)
	touch += "'" + dt + "'"
	touch += ",'" + runvars.Userid + "'"
	touch += ",'" + safesql(sqlx) + "'"
	touch += ",0)"
	_, err := DBH.Exec(touch)
	checkerr(err)
	trimHistoryLog()

}

func checkDatabaseVersion(dbx string) {

	for _, tab := range tables {
		x := getValueFromDB("SELECT name FROM sqlite_master WHERE name='"+tab+"'", "")
		if x != tab {
			fmt.Printf("ERROR: the database [%v] has no table '%v'\n", dbx, tab)
			os.Exit(1)
		}
	}

}

func trimHistoryLog() {

	maxentries := prefs.HistoryLog["maxentries"]
	if maxentries < 1 {
		return
	}
	sqlx := "SELECT count(*) AS rex FROM history"
	hrex, _ := strconv.Atoi(getValueFromDB(sqlx, "0"))
	if hrex <= maxentries {
		return
	}
	sqlx = "SELECT recordedat FROM history ORDER BY recordedat DESC LIMIT "
	sqlx += strconv.Itoa(maxentries) + ",1"
	cutoff := getValueFromDB(sqlx, "2006-01-02")
	sqlx = "DELETE FROM history WHERE recordedat < '" + cutoff + "'"
	printDebug("Trimming history - " + sqlx)
	_, err := DBH.Exec(sqlx)
	checkerr(err)

}

func check_boxes_with_contents(w http.ResponseWriter) {

	const ISODATE_NULL_LIT = "0000-00-00"

	sqlx := "SELECT boxes.boxid AS boxid"
	sqlx += ",boxes.numdocs AS box_numdocs"
	sqlx += ",boxes.min_review_date AS box_mindate"
	sqlx += ",boxes.max_review_date AS box_maxdate"
	sqlx += ",COUNT(contents.id) AS con_numdocs"
	sqlx += ",MIN(contents.review_date) AS con_mindate"
	sqlx += ",MAX(contents.review_date) AS con_maxdate"
	sqlx += " FROM contents LEFT JOIN boxes ON boxes.boxid = contents.boxid"
	sqlx += " GROUP BY contents.boxid"
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	var boxid string
	var box_numdocs, con_numdocs int
	var box_mindate, box_maxdate, con_mindate, con_maxdate string

	fmt.Fprintf(w, "<p>")
	update_commands := []string{}

	for rows.Next() {
		rows.Scan(&boxid, &box_numdocs, &box_mindate, &box_maxdate, &con_numdocs, &con_mindate, &con_maxdate)

		boxes_with_contents++

		if con_mindate == "" {
			con_mindate = ISODATE_NULL_LIT
		}
		if con_maxdate == "" {
			con_maxdate = ISODATE_NULL_LIT
		}

		if box_numdocs != con_numdocs || box_mindate != con_mindate || box_maxdate != con_maxdate {

			boxes_with_contents_updates++
			sqlx = "UPDATE boxes SET "
			sqlx += "numdocs = " + strconv.Itoa(con_numdocs)
			sqlx += ",min_review_date = '" + con_mindate + "'"
			sqlx += ",max_review_date = '" + con_maxdate + "'"
			sqlx += " WHERE boxid = '" + boxid + "'"
			//fmt.Fprintf(w, "%v<br>", sqlx)
			update_commands = append(update_commands, sqlx)
		}
	}
	fmt.Fprintln(w, "</p>")
	rows.Close()

	for _, updt := range update_commands {
		DBExec(updt)
	}
	fmt.Fprintf(w, "<h2>Database integrity checked, all ok now.</h2>")
	fmt.Fprintf(w, "<p>I checked a total of <strong>%v</strong> boxes with contents of which <strong>%v</strong> needed to be fixed.</p>", commas(boxes_with_contents), commas(boxes_with_contents_updates))

}

func check_database(w http.ResponseWriter, r *http.Request) {

	start_html(w, r)

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	//fmt.Println("Acquiring lock")
	tx, err := DBH.Begin()
	if err != nil {
		fmt.Println("Begin transaction failed")
		panic(err)
	}
	defer tx.Rollback()

	//fmt.Println("Lock acquired")
	fmt.Fprint(w, "<h2>Checking the database</h2>")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	boxes_with_contents = 0
	boxes_with_contents_updates = 0

	orphaned_boxes = []string{}
	empty_boxes = []string{}
	big_boxes = []string{}
	childless_owners = []string{}

	check_boxes_with_contents(w)
	//fmt.Println("Checked boxes with contents")
	list_orphaned_boxes()
	if len(orphaned_boxes) > 0 {
		n := len(orphaned_boxes)
		for _, box := range orphaned_boxes {
			sqlx := "UPDATE boxes SET numdocs=0 WHERE boxid='" + DBEscape(box) + "'"
			res := DBExec(sqlx)
			_, err := res.RowsAffected()
			if err != nil {
				panic(err)
			}
			//fmt.Fprintf(w, "<p>Orphaned box '%v' updated=%v</p>", orphaned_boxes, r)
		}
		zbox := `<em title="Hebrew term for a bereaved parent. Robbed of offspring, like a bear whose cubs have been taken away.&#10;&#10;Box claimed to have contents but none were found.">shakul</em>`
		if n == 1 {
			fmt.Fprintf(w, `<p>Found and zeroed one %v box</p>`, zbox)
		} else {
			fmt.Fprintf(w, `<p>Found and zeroed %v %v boxes</p>`, commas(n), zbox)
		}
	}
	//fmt.Println("Checked for zombies")

	buildOwnersTable()

	list_empty_boxes()
	if len(empty_boxes) > 0 {
		n := len(empty_boxes)
		if n == 1 {
			fmt.Fprintf(w, `<p>I found an empty box <a href="/boxes?`+Param_Labels["boxid"]+`=%v">%v</a>.</p>`, url.QueryEscape(empty_boxes[0]), empty_boxes[0])
		} else {
			fmt.Fprintf(w, `<p>I found %v empty boxes `, commas(n))
			for _, box := range empty_boxes {
				fmt.Fprintf(w, ` <a href="/boxes?`+Param_Labels["boxid"]+`=%v">%v</a> `, url.QueryEscape(box), box)
			}
			fmt.Fprintf(w, `</p>`)
		}
	}

	list_big_boxes()
	if len(big_boxes) > 0 {
		n := len(big_boxes)
		if n == 1 {
			fmt.Fprintf(w, `<p>I found a very full (&gt;%v) box <a href="/boxes?`+Param_Labels["boxid"]+`=%v">%v</a>.</p>`, prefs.MaxBoxContents, url.QueryEscape(big_boxes[0]), big_boxes[0])
		} else {
			fmt.Fprintf(w, `<p>I found %v boxes with a very large (&gt;%v) number of contents: `, commas(n), prefs.MaxBoxContents)
			for _, box := range big_boxes {
				fmt.Fprintf(w, ` <a href="/boxes?`+Param_Labels["boxid"]+`=%v">%v</a> `, url.QueryEscape(box), box)
			}
			fmt.Fprintf(w, `</p>`)
		}
	}

	list_childless_owners()
	if len(childless_owners) > 0 {
		n := len(childless_owners)
		if n == 1 {
			fmt.Fprintf(w, `<p>I found a childless owner <a href="/owners?`+Param_Labels["owner"]+`=%v">%v</a>.</p>`, url.QueryEscape(childless_owners[0]), childless_owners[0])
		} else {
			fmt.Fprintf(w, `<p>I found %v childless owners `, commas(n))
			for _, own := range childless_owners {
				fmt.Fprintf(w, ` <a href="/owners?`+Param_Labels["owner"]+`=%v">%v</a> `, url.QueryEscape(own), own)
			}
			fmt.Fprintf(w, `</p>`)
		}
	}

	n := find_duff_keys()
	if n == 1 {
		fmt.Fprint(w, "<p>One record had a key field fixed.</p>")
	} else if n > 1 {
		fmt.Fprintf(w, "<p>%v records had a key field fixed.</p>", commas(int(n)))
	}

	tx.Commit()

}

func find_duff_keys() int64 {

	update_commands := []string{}
	update_commands = append(update_commands, "UPDATE locations SET location=Trim(location) WHERE location LIKE ' %'")
	update_commands = append(update_commands, "UPDATE contents SET boxid=Trim(boxid),owner=Trim(owner),client=Trim(client) WHERE boxid LIKE ' %' OR owner LIKE ' %' OR client LIKE ' %'")
	update_commands = append(update_commands, "UPDATE boxes SET boxid=TRIM(boxid),location=Trim(location),storeref=Trim(storeref) WHERE boxid LIKE ' %' OR location LIKE ' %' OR storeref LIKE ' %'")

	var nrex int64
	for _, uc := range update_commands {
		//fmt.Println(uc)
		ucr := DBExec(uc)
		n, _ := ucr.RowsAffected()
		nrex += n
	}

	return nrex
}

func list_orphaned_boxes() {

	sqlx := "SELECT boxid,numdocs FROM boxes WHERE boxid not in (SELECT DISTINCT boxid FROM contents);"

	var boxid string
	var numdocs int
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&boxid, &numdocs)
		if numdocs > 0 {
			orphaned_boxes = append(orphaned_boxes, boxid)
		}
	}
}
func list_empty_boxes() {

	sqlx := "SELECT boxid FROM boxes WHERE numdocs=0"

	var boxid string
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&boxid)
		empty_boxes = append(empty_boxes, boxid)
	}
}

func list_childless_owners() {
	sqlx := "SELECT owners.owner FROM owners LEFT JOIN contents ON owners.owner=contents.owner WHERE contents.boxid IS NULL;"

	var owner string
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&owner)
		childless_owners = append(childless_owners, owner)
	}
}

func list_big_boxes() {

	sqlx := "SELECT boxid FROM boxes WHERE numdocs>" + strconv.Itoa(prefs.MaxBoxContents)
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	var boxid string
	for rows.Next() {
		rows.Scan(&boxid)
		big_boxes = append(big_boxes, boxid)
	}
}

func csvexp(w http.ResponseWriter, r *http.Request) {

	tab := r.FormValue(Param_Labels["table"])
	if tab == "" {
		show_search(w, r)
		return
	}
	txtname := r.FormValue(Param_Labels["textfile"])
	if txtname == "" {
		txtname = prefs.Table_Labels[tab] + ".csv"
	}
	if tab == "boxes" {
		export_boxes_csv(w, txtname)
		return
	}
	if tab == "contents" {
		export_contents_csv(w, txtname)
		return
	}
	if tab == "locations" {
		export_locations_csv(w, txtname)
		return
	}
	show_search(w, r)
}

func csvquote(x string) string {

	return `"` + strings.ReplaceAll(strings.Trim(x, " "), `"`, `""`) + `"`

}
func jsonexp(w http.ResponseWriter, r *http.Request) {

	tab := r.FormValue(Param_Labels["table"])
	if tab == "" {
		show_search(w, r)
		return
	}
	txtname := r.FormValue(Param_Labels["textfile"])
	if txtname == "" {
		txtname = prefs.Table_Labels[tab] + ".json"
	}
	if tab == "boxes" {
		export_boxes_json(w, txtname)
		return
	}
	if tab == "contents" {
		export_contents_json(w, txtname)
		return
	}
	if tab == "locations" {
		export_locations_json(w, txtname)
		return
	}
	show_search(w, r)
}

func export_boxes_csv(w http.ResponseWriter, txtname string) {

	var box table_boxes
	boxx := []string{"storeref", "boxid", "location", "overview", "numdocs", "min_review_date", "max_review_date"}

	for ix, bx := range boxx {
		boxx[ix] = strings.ReplaceAll(prefs.Field_Labels[bx], "&#8470; of ", "Num") // Yes I know
	}
	sqlx := "SELECT * FROM boxes ORDER BY boxid" // Must match cols in tablerow

	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	buffer := &bytes.Buffer{}
	writer := bufio.NewWriter(buffer)
	defer writer.Flush()

	comma := ""
	for _, x := range boxx {
		_, err = writer.WriteString(comma + `"` + x + `"`)
		checkerr(err)
		comma = ","
	}
	writer.WriteString("\r\n")

	for rows.Next() {

		rows.Scan(&box.Storeref, &box.Boxid, &box.Location, &box.Overview, &box.NumDocs, &box.Min_Review_date, &box.Max_Review_date)

		_, err = writer.WriteString(csvquote(box.Storeref) + ",")
		checkerr(err)
		_, err = writer.WriteString(csvquote(box.Boxid) + ",")
		checkerr(err)
		_, err = writer.WriteString(csvquote(box.Location) + ",")
		checkerr(err)
		_, err = writer.WriteString(csvquote(box.Overview) + ",")
		checkerr(err)
		_, err = writer.WriteString(strconv.Itoa(box.NumDocs) + ",")
		checkerr(err)
		_, err = writer.WriteString(csvquote(box.Min_Review_date) + ",")
		checkerr(err)
		_, err = writer.WriteString(csvquote(box.Max_Review_date))
		checkerr(err)

		writer.WriteString("\r\n")

	}
	writer.Flush()

	w.Header().Set("Content-Type", "text/csv") // setting the content type header to text/csv
	w.Header().Set("Content-Disposition", "attachment;filename="+txtname)
	w.Write(buffer.Bytes())

}

func export_boxes_json(w http.ResponseWriter, txtname string) {

	var box table_boxes

	w.Header().Set("Content-Type", "text/json") // setting the content type header to text/json
	w.Header().Set("Content-Disposition", "attachment;filename="+txtname)

	fmt.Fprintln(w, "[")

	sqlx := "SELECT * FROM boxes ORDER BY boxid" // Must match cols in tablerow

	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()

	commaNeeded := false
	for rows.Next() {

		rows.Scan(&box.Storeref, &box.Boxid, &box.Location, &box.Overview, &box.NumDocs, &box.Min_Review_date, &box.Max_Review_date)
		b, err := json.Marshal(box)
		if err != nil {
			panic(err)
		}
		if commaNeeded {
			fmt.Fprint(w, ",\n")
		}
		fmt.Fprint(w, string(b))
		commaNeeded = true

	}
	fmt.Fprintf(w, "\n]\n")

}

func export_contents_csv(w http.ResponseWriter, txtname string) {

	var box table_contents
	boxx := []string{"id", "boxid", "review_date", "contents", "owner", "name", "client"}

	for ix, bx := range boxx {
		boxx[ix] = strings.ReplaceAll(prefs.Field_Labels[bx], "&#8470; of ", "Num") // Yes I know
	}
	sqlx := "SELECT * FROM contents ORDER BY boxid,client" // Must match cols in tablerow

	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	buffer := &bytes.Buffer{}
	writer := bufio.NewWriter(buffer)
	defer writer.Flush()

	comma := ""
	for _, x := range boxx {
		_, err = writer.WriteString(comma + `"` + x + `"`)
		checkerr(err)
		comma = ","
	}
	writer.WriteString("\r\n")

	for rows.Next() {

		rows.Scan(&box.Id, &box.Boxid, &box.Review_date, &box.Contents, &box.Owner, &box.Name, &box.Client)

		_, err = writer.WriteString(strconv.Itoa(box.Id) + ",")
		checkerr(err)
		_, err = writer.WriteString(csvquote(box.Boxid) + ",")
		checkerr(err)
		_, err = writer.WriteString(csvquote(box.Review_date) + ",")
		checkerr(err)
		_, err = writer.WriteString(csvquote(box.Contents) + ",")
		checkerr(err)
		_, err = writer.WriteString(csvquote(box.Owner) + ",")
		checkerr(err)
		_, err = writer.WriteString(csvquote(box.Name) + ",")
		checkerr(err)
		_, err = writer.WriteString(csvquote(box.Client))
		checkerr(err)

		writer.WriteString("\r\n")

	}
	writer.Flush()

	w.Header().Set("Content-Type", "text/csv") // setting the content type header to text/csv
	w.Header().Set("Content-Disposition", "attachment;filename="+txtname)
	w.Write(buffer.Bytes())

}

func export_contents_json(w http.ResponseWriter, txtname string) {

	var box table_contents

	w.Header().Set("Content-Type", "text/json") // setting the content type header to text/json
	w.Header().Set("Content-Disposition", "attachment;filename="+txtname)

	fmt.Fprintln(w, "[")

	sqlx := "SELECT * FROM contents ORDER BY boxid,client" // Must match cols in tablerow

	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()

	commaNeeded := false
	for rows.Next() {

		rows.Scan(&box.Id, &box.Boxid, &box.Review_date, &box.Contents, &box.Owner, &box.Name, &box.Client)
		b, err := json.Marshal(box)
		if err != nil {
			panic(err)
		}
		if commaNeeded {
			fmt.Fprint(w, ",\n")
		}
		fmt.Fprint(w, string(b))
		commaNeeded = true

	}
	fmt.Fprintf(w, "\n]\n")

}

func export_locations_csv(w http.ResponseWriter, txtname string) {

	var box table_boxes
	boxx := []string{"location"}

	for ix, bx := range boxx {
		boxx[ix] = strings.ReplaceAll(prefs.Field_Labels[bx], "&#8470; of ", "Num") // Yes I know
	}
	sqlx := "SELECT location FROM locations ORDER BY location" // Must match cols in tablerow

	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	buffer := &bytes.Buffer{}
	writer := bufio.NewWriter(buffer)
	defer writer.Flush()

	for _, x := range boxx {
		_, err = writer.WriteString(`"` + x + `",`)
		if err != nil {
			panic(err)
		}
	}
	writer.WriteString("\r\n")

	for rows.Next() {

		rows.Scan(&box.Location)

		_, err = writer.WriteString(csvquote(box.Location) + ",")
		if err != nil {
			panic(err)
		}

		writer.WriteString("\r\n")

	}
	writer.Flush()

	w.Header().Set("Content-Type", "text/csv") // setting the content type header to text/csv
	w.Header().Set("Content-Disposition", "attachment;filename="+txtname)
	w.Write(buffer.Bytes())

}

func export_locations_json(w http.ResponseWriter, txtname string) {

	var box table_locations

	w.Header().Set("Content-Type", "text/json") // setting the content type header to text/json
	w.Header().Set("Content-Disposition", "attachment;filename="+txtname)

	fmt.Fprintln(w, "[")

	sqlx := "SELECT location FROM locations ORDER BY location" // Must match cols in tablerow

	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()

	commaNeeded := false
	for rows.Next() {

		rows.Scan(&box.Location)
		b, err := json.Marshal(box)
		if err != nil {
			panic(err)
		}
		if commaNeeded {
			fmt.Fprint(w, ",\n")
		}
		fmt.Fprint(w, string(b))
		commaNeeded = true

	}
	fmt.Fprintf(w, "\n]\n")

}
