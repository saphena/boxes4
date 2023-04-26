package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
type csv_boxes [7]string

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

var boxes_with_contents int
var boxes_with_contents_updates int

var orphaned_boxes []string
var duff_key_boxes []string
var empty_boxes []string
var big_boxes []string

func DBEscape(arg string) string {

	return strings.ReplaceAll(arg, "'", "''")
}

func DBExec(sqlx string) sql.Result {

	res, err := DBH.Exec(sqlx)
	if err != nil {
		fmt.Printf("DBExec = %v\n", sqlx)
		panic(err)
	}
	return res

}

func check_boxes_with_contents(w http.ResponseWriter, r *http.Request) {

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
	if err != nil {
		panic(err)
	}
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
			fmt.Fprintf(w, "%v<br>", sqlx)
			update_commands = append(update_commands, sqlx)
		}
	}
	fmt.Fprintln(w, "</p>")
	rows.Close()

	for _, updt := range update_commands {
		DBExec(updt)
	}
	fmt.Fprintf(w, "<h2>Database integrity checked, all ok now.</h2>")
	fmt.Fprintf(w, "<p>I checked a total of <strong>%v</strong> boxes with contents of which <strong>%v</strong> needed to be fixed.</p>", boxes_with_contents, boxes_with_contents_updates)

}

func check_database(w http.ResponseWriter, r *http.Request) {

	start_html(w, r)

	//fmt.Println("Acquiring lock")
	tx, err := DBH.Begin()
	if err != nil {
		fmt.Println("Begin transaction failed")
		panic(err)
	}
	defer tx.Rollback()

	//fmt.Println("Lock acquired")
	fmt.Fprint(w, "<h2>Checking the database</h2>")

	boxes_with_contents = 0
	boxes_with_contents_updates = 0

	orphaned_boxes = []string{}
	duff_key_boxes = []string{}
	empty_boxes = []string{}
	big_boxes = []string{}

	check_boxes_with_contents(w, r)
	//fmt.Println("Checked boxes with contents")
	list_orphaned_boxes()
	if len(orphaned_boxes) > 0 {
		n := len(orphaned_boxes)
		for _, box := range orphaned_boxes {
			sqlx := "UPDATE boxes SET numdocs=0 WHERE boxid='" + DBEscape(box) + "'"
			res := DBExec(sqlx)
			r, err := res.RowsAffected()
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(w, "<p>Orphaned box '%v' updated=%v</p>", orphaned_boxes, r)
		}
		if n == 1 {
			fmt.Fprint(w, `<p>Found and zeroed one zombie box</p>`)
		} else {
			fmt.Fprintf(w, `<p>Found and zeroed %v zombie boxes</p>`, n)
		}
	}
	//fmt.Println("Checked for zombies")

	list_empty_boxes()
	if len(empty_boxes) > 0 {
		n := len(empty_boxes)
		if n == 1 {
			fmt.Fprintf(w, `<p>I found an empty box <a href="/boxes?`+Param_Labels["boxid"]+`=%v">%v</a>.</p>`, url.QueryEscape(empty_boxes[0]), empty_boxes[0])
		} else {
			fmt.Fprintf(w, `<p>I found %v empty boxes `, n)
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
			fmt.Fprintf(w, `<p>I found a very full (&gt;%v) box <a href="/boxes?`+Param_Labels["boxid"]+`=%v">%v</a>.</p>`, MAX_BOX_CONTENTS, url.QueryEscape(big_boxes[0]), big_boxes[0])
		} else {
			fmt.Fprintf(w, `<p>I found %v boxes with a very large (&gt;%v) number of contents: `, n, MAX_BOX_CONTENTS)
			for _, box := range big_boxes {
				fmt.Fprintf(w, ` <a href="/boxes?`+Param_Labels["boxid"]+`=%v">%v</a> `, url.QueryEscape(box), box)
			}
			fmt.Fprintf(w, `</p>`)
		}
	}

	find_duff_keys()
	if len(duff_key_boxes) > 0 {
		n := len(duff_key_boxes)
		txt := "with files having a space before either partner initials or client number"
		if n == 1 {
			fmt.Fprintf(w, `<p>I found one box <a href="/boxes?`+Param_Labels["boxid"]+`=%v">%v</a> %v.</p>`, url.QueryEscape(duff_key_boxes[0]), duff_key_boxes[0], txt)
		} else {
			fmt.Fprintf(w, `<p>I found %v boxes %v:`, n, txt)
			for _, box := range duff_key_boxes {
				fmt.Fprintf(w, ` <a href="/boxes?`+Param_Labels["boxid"]+`=%v">%v</a> `, url.QueryEscape(box), box)
			}
			fmt.Fprintf(w, `</p>`)
		}
	}

	tx.Commit()

}

func find_duff_keys() {

	sqlx := "SELECT boxid FROM contents WHERE owner LIKE ' %' OR client LIKE ' %'"
	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var boxid string
	for rows.Next() {
		rows.Scan(&boxid)
		if !contains(duff_key_boxes, boxid) {
			duff_key_boxes = append(duff_key_boxes, boxid)
		}
	}
}

func list_orphaned_boxes() {

	sqlx := "SELECT boxid,numdocs FROM boxes WHERE boxid not in (SELECT DISTINCT boxid FROM contents);"

	var boxid string
	var numdocs int
	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
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
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&boxid)
		empty_boxes = append(empty_boxes, boxid)
	}
}

func list_big_boxes() {

	sqlx := "SELECT boxid FROM boxes WHERE numdocs>" + strconv.Itoa(MAX_BOX_CONTENTS)
	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
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
	if tab == "boxes" {
		export_boxes_csv(w)
		return
	}
	if tab == "contents" {
		export_contents_csv(w)
		return
	}
	if tab == "locations" {
		export_locations_csv(w)
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
	if tab == "boxes" {
		export_boxes_json(w)
		return
	}
	if tab == "contents" {
		export_contents_json(w)
		return
	}
	if tab == "locations" {
		export_locations_json(w)
		return
	}
	show_search(w, r)
}

func export_boxes_csv(w http.ResponseWriter) {

	var box table_boxes
	boxx := []string{"storeref", "boxid", "location", "overview", "numdocs", "min_review_date", "max_review_date"}

	for ix, bx := range boxx {
		boxx[ix] = strings.ReplaceAll(Field_Labels[bx], "&#8470; of ", "Num") // Yes I know
	}
	sqlx := "SELECT * FROM boxes ORDER BY boxid" // Must match cols in tablerow

	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
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

		rows.Scan(&box.Storeref, &box.Boxid, &box.Location, &box.Overview, &box.NumDocs, &box.Min_Review_date, &box.Max_Review_date)

		_, err = writer.WriteString(csvquote(box.Storeref) + ",")
		if err != nil {
			panic(err)
		}
		_, err = writer.WriteString(csvquote(box.Boxid) + ",")
		if err != nil {
			panic(err)
		}
		_, err = writer.WriteString(csvquote(box.Location) + ",")
		if err != nil {
			panic(err)
		}
		_, err = writer.WriteString(csvquote(box.Overview) + ",")
		if err != nil {
			panic(err)
		}
		_, err = writer.WriteString(strconv.Itoa(box.NumDocs) + ",")
		if err != nil {
			panic(err)
		}
		_, err = writer.WriteString(csvquote(box.Min_Review_date) + ",")
		if err != nil {
			panic(err)
		}
		_, err = writer.WriteString(csvquote(box.Max_Review_date))
		if err != nil {
			panic(err)
		}

		writer.WriteString("\r\n")

	}
	writer.Flush()

	w.Header().Set("Content-Type", "text/csv") // setting the content type header to text/csv
	w.Header().Set("Content-Disposition", "attachment;filename=boxes.csv")
	w.Write(buffer.Bytes())

}

func export_boxes_json(w http.ResponseWriter) {

	var box table_boxes

	w.Header().Set("Content-Type", "text/json") // setting the content type header to text/json
	w.Header().Set("Content-Disposition", "attachment;filename=boxes.json")

	fmt.Fprintln(w, "[")

	sqlx := "SELECT * FROM boxes ORDER BY boxid" // Must match cols in tablerow

	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
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

func export_contents_csv(w http.ResponseWriter) {

	var box table_contents
	boxx := []string{"id", "boxid", "review_date", "contents", "owner", "name", "client"}

	for ix, bx := range boxx {
		boxx[ix] = strings.ReplaceAll(Field_Labels[bx], "&#8470; of ", "Num") // Yes I know
	}
	sqlx := "SELECT * FROM contents ORDER BY boxid,client" // Must match cols in tablerow

	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
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

		rows.Scan(&box.Id, &box.Boxid, &box.Review_date, &box.Contents, &box.Owner, &box.Name, &box.Client)

		_, err = writer.WriteString(strconv.Itoa(box.Id) + ",")
		if err != nil {
			panic(err)
		}
		_, err = writer.WriteString(csvquote(box.Boxid) + ",")
		if err != nil {
			panic(err)
		}
		_, err = writer.WriteString(csvquote(box.Review_date) + ",")
		if err != nil {
			panic(err)
		}
		_, err = writer.WriteString(csvquote(box.Contents) + ",")
		if err != nil {
			panic(err)
		}
		_, err = writer.WriteString(csvquote(box.Owner) + ",")
		if err != nil {
			panic(err)
		}
		_, err = writer.WriteString(csvquote(box.Name) + ",")
		if err != nil {
			panic(err)
		}
		_, err = writer.WriteString(csvquote(box.Client))
		if err != nil {
			panic(err)
		}

		writer.WriteString("\r\n")

	}
	writer.Flush()

	w.Header().Set("Content-Type", "text/csv") // setting the content type header to text/csv
	w.Header().Set("Content-Disposition", "attachment;filename=contents.csv")
	w.Write(buffer.Bytes())

}

func export_contents_json(w http.ResponseWriter) {

	var box table_contents

	w.Header().Set("Content-Type", "text/json") // setting the content type header to text/json
	w.Header().Set("Content-Disposition", "attachment;filename=contents.json")

	fmt.Fprintln(w, "[")

	sqlx := "SELECT * FROM contents ORDER BY boxid,client" // Must match cols in tablerow

	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
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

func export_locations_csv(w http.ResponseWriter) {

	var box table_boxes
	boxx := []string{"location"}

	for ix, bx := range boxx {
		boxx[ix] = strings.ReplaceAll(Field_Labels[bx], "&#8470; of ", "Num") // Yes I know
	}
	sqlx := "SELECT location FROM locations ORDER BY location" // Must match cols in tablerow

	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
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
	w.Header().Set("Content-Disposition", "attachment;filename=locations.csv")
	w.Write(buffer.Bytes())

}

func export_locations_json(w http.ResponseWriter) {

	var box table_locations

	w.Header().Set("Content-Type", "text/json") // setting the content type header to text/json
	w.Header().Set("Content-Disposition", "attachment;filename=locations.json")

	fmt.Fprintln(w, "[")

	sqlx := "SELECT location FROM locations ORDER BY location" // Must match cols in tablerow

	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
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