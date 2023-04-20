package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

func showbox(w http.ResponseWriter, r *http.Request) {

	if r.FormValue(Param_Labels["boxid"]) == "" {
		show_search(w, r)
		return
	}

	start_html(w)

	sqlboxid := strings.ReplaceAll(r.FormValue(Param_Labels["boxid"]), "'", "''")
	sqlx := "SELECT storeref,boxid,location,overview,numdocs,min_review_date,max_review_date FROM boxes WHERE boxid='" + sqlboxid + "'"
	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var bv boxvars
	bv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])

	if !rows.Next() {
		fmt.Fprintf(w, "<p>Bugger! %v</p>", r.FormValue(Param_Labels["boxid"]))
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
	showBoxfiles(w, r, sqlboxid)

}

func showBoxfiles(w http.ResponseWriter, r *http.Request, boxid string) {

	NumFiles, _ := strconv.Atoi(getValueFromDB("SELECT COUNT(*) AS rex FROM contents WHERE boxid='"+boxid+"'", "rex", "0"))
	sqllimit := emit_page_anchors(w, r, "boxes", NumFiles)
	sqlx := "SELECT owner,client,name,contents,review_date FROM contents WHERE boxid='" + boxid + "'"

	if r.FormValue(Param_Labels["order"]) != "" {
		sqlx += " ORDER BY TRIM(contents." + r.FormValue(Param_Labels["order"]) + ")"
		if r.FormValue(Param_Labels["desc"]) != "" {
			sqlx += " DESC"
		}
	} else {
		sqlx += " ORDER BY owner,client"
	}
	rows, _ := DBH.Query(sqlx + sqllimit)
	defer rows.Close()

	html, err := template.New("").Parse(boxfileshdr)
	if err != nil {
		panic(err)
	}

	var bfv boxfilevars
	bfv.Boxid = boxid
	bfv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])

	err = html.Execute(w, bfv)
	if err != nil {
		panic(err)
	}

	nrows := 0
	html, err = template.New("").Parse(boxfilesline)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		rows.Scan(&bfv.Owner, &bfv.Client, &bfv.Name, &bfv.Contents, &bfv.Date)
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
