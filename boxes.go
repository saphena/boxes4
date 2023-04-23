package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func showboxes(w http.ResponseWriter, r *http.Request) {

	if r.FormValue(Param_Labels["boxid"]) != "" {
		showbox(w, r)
		return
	}

	start_html(w, r)

	sqlx := " FROM boxes "

	NumBoxes, _ := strconv.Atoi(getValueFromDB("SELECT Count(*) As rex "+sqlx, "rex", "0"))

	if r.FormValue(Param_Labels["order"]) != "" {
		sqlx += "ORDER BY " + r.FormValue(Param_Labels["order"])
		if r.FormValue(Param_Labels["desc"]) != "" {
			sqlx += " DESC"
		}
	} else {
		sqlx += "ORDER BY boxid"
	}

	flds := " storeref,boxid,location,overview,numdocs,min_review_date,max_review_date "
	sqlx += emit_page_anchors(w, r, "boxes", NumBoxes)
	rows, err := DBH.Query("SELECT  " + flds + sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var box boxvars
	box.Single = r.FormValue(Param_Labels["boxid"]) != ""
	box.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"]) || r.FormValue(Param_Labels["order"]) == ""

	html, err := template.New("").Parse(boxtablehdr)
	if err != nil {
		panic(err)
	}
	err = html.Execute(w, box)
	if err != nil {
		panic(err)
	}

	html, err = template.New("").Parse(boxtablerow)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		rows.Scan(&box.Storeref, &box.Boxid, &box.Location, &box.Overview, &box.NumFiles, &box.Min_review_date, &box.Max_review_date)
		box.StorerefUrl = template.URLQueryEscaper(box.Storeref)
		box.BoxidUrl = template.URLQueryEscaper(box.Boxid)
		box.LocationUrl = template.URLQueryEscaper(box.Location)
		if box.Max_review_date == box.Min_review_date {
			box.Date = box.Max_review_date
			box.Single = true
		} else {
			box.Date = box.Min_review_date + " to " + box.Max_review_date
			box.Single = false
		}
		err := html.Execute(w, box)
		if err != nil {
			panic(err)
		}
	}
	fmt.Fprint(w, ownerlisttrailer)

}

func showbox(w http.ResponseWriter, r *http.Request) {

	if r.FormValue(Param_Labels["boxid"]) == "" {
		show_search(w, r)
		return
	}

	start_html(w, r)

	sqlboxid, _ := url.QueryUnescape(r.FormValue(Param_Labels["boxid"]))
	sqlboxid = strings.ReplaceAll(sqlboxid, "'", "''")
	sqlx := "SELECT storeref,boxid,location,overview,numdocs,min_review_date,max_review_date FROM boxes WHERE boxid='" + sqlboxid + "'"
	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var bv boxvars
	bv.Single = r.FormValue(Param_Labels["boxid"]) != ""
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
		bfv.OwnerUrl = template.URLQueryEscaper(bfv.Owner)
		bfv.ClientUrl = template.URLQueryEscaper(bfv.Client)
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
