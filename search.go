package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

func exec_search(w http.ResponseWriter, r *http.Request) {

	start_html(w)

	var sqlx = ` FROM contents LEFT JOIN boxes ON contents.boxid=boxes.boxid `
	if r.FormValue(Param_Labels["find"]) != "" {
		sqlx += `WHERE ((contents.boxid = '?')
		OR (boxes.storeref = '?') 
		OR (boxes.overview LIKE '%?%')
        OR (contents.owner = '?') 
        OR (contents.client = '?') 
        OR (contents.contents LIKE '%?%') 
        OR (contents.name LIKE '%?%')) 

		`
	}
	sqlx = strings.ReplaceAll(sqlx, "?", strings.ReplaceAll(r.FormValue(Param_Labels["find"]), "'", "''"))
	if r.FormValue(Param_Labels["order"]) != "" {
		sqlx += " ORDER BY TRIM(contents." + r.FormValue(Param_Labels["order"]) + ")"
		if r.FormValue(Param_Labels["desc"]) != "" {
			sqlx += " DESC"
		}
	}

	// fmt.Println(sqlx)

	FoundRecCount, _ := strconv.Atoi(getValueFromDB("SELECT Count(*) AS Rexx"+sqlx, "Rexx", "0"))

	var res searchResultsVar
	res.Boxid = order_dir(r, "boxid")
	res.Owner = order_dir(r, "owner")
	res.Client = order_dir(r, "client")
	res.Name = order_dir(r, "name")
	res.Date = order_dir(r, "review_date")
	res.Find = r.FormValue(Param_Labels["find"])
	res.Found = strconv.Itoa(FoundRecCount)

	html, err := template.New("main").Parse(searchResultsHdr1)
	if err != nil {
		panic(err)
	}
	html.Execute(w, res)

	flds := "contents.BoxID,contents.Owner,contents.Client,contents.Name,contents.Contents,contents.Review_Date"

	sqllimit := emit_page_anchors(w, r, "find", FoundRecCount)
	rows, err := DBH.Query("SELECT " + flds + sqlx + sqllimit)
	if err != nil {
		fmt.Printf("Omg! %v\n", sqlx)
		panic(err)
	}
	html, err = template.New("main").Parse(searchResultsHdr2)
	if err != nil {
		panic(err)
	}
	html.Execute(w, res)

	for rows.Next() {
		rows.Scan(&res.Boxid, &res.Owner, &res.Client, &res.Name, &res.Contents, &res.Date)
		html, _ = template.New("main").Parse(searchResultsLine)
		html.Execute(w, res)
	}
	html, _ = template.New("main").Parse(searchResultsTrailer)
	html.Execute(w, "")

}

func show_search(w http.ResponseWriter, r *http.Request) {

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
