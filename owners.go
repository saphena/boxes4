package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func showowners(w http.ResponseWriter, r *http.Request) {

	start_html(w)

	owner, _ := url.QueryUnescape(r.FormValue(Param_Labels["owner"]))
	sqlx := "SELECT DISTINCT TRIM(owner), COUNT(TRIM(owner)) AS numdocs FROM contents "
	sqlx += "GROUP BY TRIM(owner) "
	if owner != "" {
		sqlx += "HAVING TRIM(owner) = '" + strings.ReplaceAll(owner, "'", "''") + "' "
	}

	if r.FormValue(Param_Labels["order"]) != "" {
		sqlx += "ORDER BY " + r.FormValue(Param_Labels["order"])
		if r.FormValue(Param_Labels["desc"]) != "" {
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

	sqllimit := ""
	if owner == "" {
		sqllimit = emit_page_anchors(w, r, "owners", nrex)
	}
	rows, err = DBH.Query(sqlx + sqllimit)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var plv ownerlistvars
	plv.Single = owner != ""

	html, err := template.New("").Parse(ownerlisthdr)
	if err != nil {
		panic(err)
	}
	plv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])
	plv.NumOrder = r.FormValue(Param_Labels["order"]) == Param_Labels["numdocs"]
	html.Execute(w, plv)

	html, err = template.New("").Parse(ownerlistline)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		rows.Scan(&plv.Owner, &plv.NumFiles)
		plv.OwnerUrl = template.URLQueryEscaper(plv.Owner)
		err := html.Execute(w, plv)
		if err != nil {
			panic(err)
		}
	}
	fmt.Fprint(w, ownerlisttrailer)

	if owner == "" {
		return
	}

	rows.Close()

	sqlx = " FROM contents WHERE owner='" + strings.ReplaceAll(owner, "'", "''") + "'"
	NumRows, _ := strconv.Atoi(getValueFromDB("SELECT COUNT(*) AS rex"+sqlx, "rex", "0"))
	if r.FormValue(Param_Labels["order"]) != "" {
		sqlx += " ORDER BY Upper(Trim(contents." + r.FormValue(Param_Labels["order"]) + "))"
		if r.FormValue(Param_Labels["desc"]) != "" {
			sqlx += " DESC"
		}
	}

	sqllimit = emit_page_anchors(w, r, "owners", NumRows)

	rows, err = DBH.Query("SELECT boxid,client,name,contents,review_date " + sqlx + sqllimit)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var ofv ownerfilesvar
	ofv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])
	ofv.Owner = owner
	ofv.OwnerUrl = template.URLQueryEscaper(ofv.Owner)
	html, err = template.New("").Parse(ownerfileshdr)
	if err != nil {
		panic(err)
	}
	err = html.Execute(w, ofv)
	if err != nil {
		panic(err)
	}

	html, err = template.New("").Parse(ownerfilesline)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		rows.Scan(&ofv.Boxid, &ofv.Client, &ofv.Name, &ofv.Contents, &ofv.Date)
		ofv.BoxidUrl = template.URLQueryEscaper(ofv.Boxid)
		ofv.ClientUrl = template.URLQueryEscaper(ofv.Client)
		err = html.Execute(w, ofv)
	}
	fmt.Fprint(w, ownerfilestrailer)
}
