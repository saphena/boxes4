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

	var ownerlisthdr = `
<table class="ownerlist">
<thead>
<tr>


<th class="owner">{{if .Single}}{{else}}<a title="&#8645;" href="/owners?` + Param_Labels["order"] + `=owner{{if .Desc}}&` + Param_Labels["desc"] + `=owner{{end}}">{{end}}` + prefs.Field_Labels["owner"] + `{{if .Single}}{{else}}</a>{{end}}</th>
<th class="number">{{if .Single}}{{else}}<a title="&#8645;" href="/owners?` + Param_Labels["order"] + `=numdocs{{if .Desc}}&` + Param_Labels["desc"] + `=numdocs{{end}}">{{end}}` + prefs.Field_Labels["numdocs"] + `{{if .Single}}{{else}}</a>{{end}}</th>
</tr>
</thead>
<tbody>
`

	var ownerfileshdr = `
<table class="ownerfiles">
<thead>
<tr>

<th class="owner"><a title="&#8645;" href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=boxid{{if .Desc}}&` + Param_Labels["desc"] + `=boxid{{end}}">` + prefs.Field_Labels["boxid"] + `</a></th>
<th class="client"><a title="&#8645;" href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=client{{if .Desc}}&` + Param_Labels["desc"] + `=client{{end}}">` + prefs.Field_Labels["client"] + `</a></th>
<th class="name"><a title="&#8645;" href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=name{{if .Desc}}&` + Param_Labels["desc"] + `=name{{end}}">` + prefs.Field_Labels["name"] + `</a></th>
<th class="contents"><a title="&#8645;" href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=contents{{if .Desc}}&` + Param_Labels["desc"] + `=contents{{end}}">` + prefs.Field_Labels["contents"] + `</a></th>
<th class="review_date"><a title="&#8645;" href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=review_date{{if .Desc}}&` + Param_Labels["desc"] + `=review_date{{end}}">` + prefs.Field_Labels["review_date"] + `</a></th>
</tr>
</thead>
<tbody>
`

	start_html(w, r)

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
	checkerr(err)
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
	checkerr(err)
	defer rows.Close()

	var plv ownerlistvars
	plv.Single = owner != ""

	html, err := template.New("").Parse(ownerlisthdr)
	checkerr(err)
	plv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])
	plv.NumOrder = r.FormValue(Param_Labels["order"]) == Param_Labels["numdocs"]
	html.Execute(w, plv)

	html, err = template.New("").Parse(ownerlistline)
	checkerr(err)
	for rows.Next() {
		rows.Scan(&plv.Owner, &plv.NumFiles)
		plv.OwnerUrl = template.URLQueryEscaper(plv.Owner)
		plv.NumFilesX = commas(plv.NumFiles)
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

	sqlx = " FROM contents  LEFT JOIN boxes ON contents.boxid=boxes.boxid "
	sqlx += " WHERE owner='" + strings.ReplaceAll(owner, "'", "''") + "'"
	NumRows, _ := strconv.Atoi(getValueFromDB("SELECT COUNT(*) AS rex"+sqlx, "rex", "0"))
	if r.FormValue(Param_Labels["order"]) != "" {
		sqlx += " ORDER BY Upper(Trim(contents." + r.FormValue(Param_Labels["order"]) + "))"
		if r.FormValue(Param_Labels["desc"]) != "" {
			sqlx += " DESC"
		}
	}

	sqllimit = emit_page_anchors(w, r, "owners", NumRows)

	rows, err = DBH.Query("SELECT contents.boxid,client,name,contents,review_date,overview " + sqlx + sqllimit)
	checkerr(err)
	defer rows.Close()
	var ofv ownerfilesvar
	ofv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])
	ofv.Owner = owner
	ofv.OwnerUrl = template.URLQueryEscaper(ofv.Owner)
	html, err = template.New("").Parse(ownerfileshdr)
	checkerr(err)
	err = html.Execute(w, ofv)
	checkerr(err)

	html, err = template.New("").Parse(ownerfilesline)
	checkerr(err)
	for rows.Next() {
		rows.Scan(&ofv.Boxid, &ofv.Client, &ofv.Name, &ofv.Contents, &ofv.Date, &ofv.Overview)
		ofv.BoxidUrl = template.URLQueryEscaper(ofv.Boxid)
		ofv.ClientUrl = template.URLQueryEscaper(ofv.Client)
		err = html.Execute(w, ofv)
	}
	fmt.Fprint(w, ownerfilestrailer)
}
