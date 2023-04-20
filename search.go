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
		if r.FormValue(Param_Labels["field"]) != "" {
			sqlx += `WHERE ` + r.FormValue((Param_Labels["field"])) + `= '?'`
		} else {
			sqlx += `WHERE ((contents.boxid = '?')
		OR (boxes.storeref = '?') 
		OR (boxes.overview LIKE '%?%')
        OR (contents.owner = '?') 
        OR (contents.client = '?') 
        OR (contents.contents LIKE '%?%') 
        OR (contents.name LIKE '%?%')) 
		OR (contents.review_date = '?')
		`
		}
	}
	sqlx = strings.ReplaceAll(sqlx, "?", strings.ReplaceAll(r.FormValue(Param_Labels["find"]), "'", "''"))
	if r.FormValue(Param_Labels["order"]) != "" {
		sqlx += " ORDER BY Upper(Trim(contents." + r.FormValue(Param_Labels["order"]) + "))"
		if r.FormValue(Param_Labels["desc"]) != "" {
			sqlx += " DESC"
		}
	}

	FoundRecCount, _ := strconv.Atoi(getValueFromDB("SELECT Count(*) AS Rexx"+sqlx, "Rexx", "0"))

	var res searchResultsVar

	res.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])

	res.Boxid = order_dir(r, "boxid")
	res.Owner = order_dir(r, "owner")
	res.Client = order_dir(r, "client")
	res.Name = order_dir(r, "name")
	res.Date = order_dir(r, "review_date")
	res.Find = r.FormValue(Param_Labels["find"])
	res.Found = strconv.Itoa(FoundRecCount)
	res.Field = Field_Labels[r.FormValue(Param_Labels["field"])]

	html, err := template.New("main").Parse(searchResultsHdr1)
	if err != nil {
		panic(err)
	}
	html.Execute(w, res)

	flds := "contents.boxid,contents.owner,contents.client,contents.name,contents.contents,contents.review_date,boxes.storeref,boxes.overview"

	sqllimit := emit_page_anchors(w, r, "find", FoundRecCount)

	//fmt.Printf("DEBUG: sql = SELECT %v%v%v\n", flds, sqlx, sqllimit)

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

	html, err = template.New("main").Parse(searchResultsLine)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		rows.Scan(&res.Boxid, &res.Owner, &res.Client, &res.Name, &res.Contents, &res.Date, &res.Storeref, &res.Overview)
		err = html.Execute(w, res)
		if err != nil {
			panic(err)
		}
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
