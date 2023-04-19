package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

func showowners(w http.ResponseWriter, r *http.Request) {

	start_html(w)

	sqlx := "SELECT DISTINCT TRIM(owner), COUNT(TRIM(owner)) AS numdocs FROM contents "
	sqlx += "GROUP BY TRIM(owner) "
	if r.FormValue(Param_Labels["owner"]) != "" {
		sqlx += "HAVING TRIM(owner) = '" + r.FormValue(Param_Labels["owner"]) + "' "
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
	if r.FormValue(Param_Labels["owner"]) == "" {
		sqllimit = emit_page_anchors(w, r, "partners", nrex)
	}
	rows, err = DBH.Query(sqlx + sqllimit)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var plv partnerlistvars

	html, err := template.New("").Parse(partnerlisthdr)
	if err != nil {
		panic(err)
	}
	plv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])
	plv.NumOrder = r.FormValue(Param_Labels["order"]) == Param_Labels["numdocs"]
	html.Execute(w, plv)

	html, err = template.New("").Parse(partnerlistline)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		rows.Scan(&plv.Owner, &plv.NumFiles)
		err := html.Execute(w, plv)
		if err != nil {
			panic(err)
		}
	}
	fmt.Fprint(w, partnerlisttrailer)

	if r.FormValue(Param_Labels["owner"]) == "" {
		return
	}

	rows.Close()

	sqlx = " FROM contents WHERE owner='" + strings.ReplaceAll(r.FormValue(Param_Labels["owner"]), "'", "''") + "'"
	NumRows, _ := strconv.Atoi(getValueFromDB("SELECT COUNT(*) AS rex"+sqlx, "rex", "0"))

	sqllimit = emit_page_anchors(w, r, "partners", NumRows)

	rows, err = DBH.Query("SELECT * " + sqlx + sqllimit)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

}
