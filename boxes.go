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

	var boxtablehdr = `
<table class="boxlist">
<thead>
<tr>


<th class="boxid"><a title="&#8645;" href="/boxes?` + Param_Labels["order"] + `=boxid{{if .Desc}}&` + Param_Labels["desc"] + `=boxid{{end}}">` + prefs.Field_Labels["boxid"] + `</a></th>
<th class="location"><a title="&#8645;" href="/boxes?` + Param_Labels["order"] + `=location{{if .Desc}}&` + Param_Labels["desc"] + `=location{{end}}">` + prefs.Field_Labels["location"] + `</a></th>
<th class="storeref"><a title="&#8645;" href="/boxes?` + Param_Labels["order"] + `=storeref{{if .Desc}}&` + Param_Labels["desc"] + `=storeref{{end}}">` + prefs.Field_Labels["storeref"] + `</a></th>
<th class="contents"><a title="&#8645;" href="/boxes?` + Param_Labels["order"] + `=overview{{if .Desc}}&` + Param_Labels["desc"] + `=overview{{end}}">` + prefs.Field_Labels["overview"] + `</a></th>
<th class="boxid"><a title="&#8645;" href="/boxes?` + Param_Labels["order"] + `=numdocs{{if .Desc}}&` + Param_Labels["desc"] + `=numdocs{{end}}">` + prefs.Field_Labels["numdocs"] + `</a></th>
<th class="boxid"><a title="&#8645;" href="/boxes?` + Param_Labels["order"] + `=min_review_date{{if .Desc}}&` + Param_Labels["desc"] + `=min_review_date{{end}}">` + prefs.Field_Labels["review_date"] + `</a></th>
</tr>
</thead>
<tbody>
`

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
		box.NumFilesX = commas(box.NumFiles)
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

	var boxhtml = `
<table class="boxheader">


<tr><td class="vlabel">{{if .Single}}{{else}}<a title="&#8645;" href="/boxes?` + Param_Labels["boxid"] + `={{.BoxidUrl}}&` + Param_Labels["order"] + `=boxid&` + Param_Labels["desc"] + `=boxid">{{end}}` + prefs.Field_Labels["boxid"] + `{{if .Single}}{{else}}</a>{{end}} : </td><td class="vdata">{{.Boxid}}</td></tr>
<tr><td class="vlabel">` + prefs.Field_Labels["location"] + ` : </td><td class="vdata"><a href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}">{{.Location}}</a></td></tr>
<tr><td class="vlabel">` + prefs.Field_Labels["storeref"] + ` : </td><td class="vdata"><a href="/find?` + Param_Labels["find"] + `={{.StorerefUrl}}&` + Param_Labels["field"] + `=storeref">{{.Storeref}}</a></td></tr>
<tr><td class="vlabel">` + prefs.Field_Labels["contents"] + ` : </td><td class="vdata">{{.Contents}}</td></tr>
<tr><td class="vlabel">` + prefs.Field_Labels["numdocs"] + ` : </td><td class="vdata numdocs">{{.NumFilesX}}</td></tr>
<tr><td class="vlabel">` + prefs.Field_Labels["review_date"] + ` : </td><td class="vdata">{{.Date}}</td></tr>

</table>
`

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
	xx := prefs.Field_Labels["boxid"]
	fmt.Printf("xx=%v\n", xx)
	var mindate, maxdate string
	rows.Scan(&bv.Storeref, &bv.Boxid, &bv.Location, &bv.Contents, &bv.NumFiles, &mindate, &maxdate)
	bv.Date = mindate + " to " + maxdate
	bv.NumFilesX = commas(bv.NumFiles)
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

	var boxfileshdr = `
<table class="boxfiles">
<thead>
<tr>
<th class="owner"><a title="&#8645;" href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=owner{{if .Desc}}&` + Param_Labels["desc"] + `=owner{{end}}">` + prefs.Field_Labels["owner"] + `</a></th>
<th class="owner"><a title="&#8645;" href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=client{{if .Desc}}&` + Param_Labels["desc"] + `=client{{end}}">` + prefs.Field_Labels["client"] + `</a></th>
<th class="owner"><a title="&#8645;" href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=name{{if .Desc}}&` + Param_Labels["desc"] + `=name{{end}}">` + prefs.Field_Labels["name"] + `</a></th>
<th class="owner"><a title="&#8645;" href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=contents{{if .Desc}}&` + Param_Labels["desc"] + `=contents{{end}}">` + prefs.Field_Labels["contents"] + `</a></th>
<th class="owner"><a href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=review_date{{if .Desc}}&` + Param_Labels["desc"] + `=review_date{{end}}">` + prefs.Field_Labels["review_date"] + `</a></th>


</tr>
</thead>
<tbody>
`

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
