package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func showlocations(w http.ResponseWriter, r *http.Request) {

	var locationlisthdr = `
<table class="locationlist">
<thead>
<tr>


<th class="location">{{if .Single}}{{else}}<a title="&#8645;" href="/locations?` + Param_Labels["order"] + `=location{{if .Desc}}&` + Param_Labels["desc"] + `=location{{end}}">{{end}}` + prefs.Field_Labels["location"] + `{{if .Single}}{{else}}</a>{{end}}</th>
<th class="numboxes">{{if .Single}}{{else}}<a title="&#8645;" href="/locations?` + Param_Labels["order"] + `=numboxes{{if .Desc}}&` + Param_Labels["desc"] + `=numboxes{{end}}">{{end}}` + prefs.Field_Labels["numboxes"] + `{{if .Single}}{{else}}</a>{{end}}</th>
</tr>
</thead>
<tbody>
`

	start_html(w, r)

	sqlx := " FROM locations "

	NumLocations, _ := strconv.Atoi(getValueFromDB("SELECT Count(*) As rex "+sqlx, "rex", "0"))

	sqlx = " FROM locations RIGHT JOIN boxes ON locations.location=boxes.location"

	sqllocation := ""
	if r.FormValue(Param_Labels["location"]) != "" {
		sqllocation, _ = url.QueryUnescape(r.FormValue(Param_Labels["location"]))
		sqllocation = strings.ReplaceAll(sqllocation, "'", "''")
		sqlx += " WHERE locations.location = '" + sqllocation + "'"

	}

	sqlx += " GROUP BY locations.location "
	if r.FormValue(Param_Labels["order"]) != "" && r.FormValue(Param_Labels["location"]) == "" {
		sqlx += "ORDER BY locations." + r.FormValue(Param_Labels["order"])
		if r.FormValue(Param_Labels["desc"]) != "" {
			sqlx += " DESC"
		}
	} else {
		sqlx += "ORDER BY locations.location"
	}

	flds := " id,locations.location, Count(boxid) As NumBoxes "
	if r.FormValue(Param_Labels["location"]) == "" {
		sqlx += emit_page_anchors(w, r, "locations", NumLocations)
	}
	//fmt.Printf("DEBUG: SELECT %v%v\n", flds, sqlx)
	rows, err := DBH.Query("SELECT  " + flds + sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var loc locationlistvars
	loc.Single = r.FormValue(Param_Labels["location"]) != ""
	loc.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"]) || r.FormValue(Param_Labels["order"]) == ""

	temp, err := template.New("locationlisthdr").Parse(locationlisthdr)
	if err != nil {
		panic(err)
	}
	err = temp.Execute(w, loc)
	if err != nil {
		panic(err)
	}

	if true {
		temp, err = template.New("locationlistline2").Parse(locationlistline)
		if err != nil {
			panic(err)
		}
		for rows.Next() {
			rows.Scan(&loc.Id, &loc.Location, &loc.NumBoxes)
			loc.LocationUrl = url.QueryEscape(loc.Location)
			loc.NumBoxesX = commas(loc.NumBoxes)
			err := temp.Execute(w, loc)
			if err != nil {
				panic(err)
			}
		}
		fmt.Fprint(w, ownerlisttrailer)
	}
	if sqllocation != "" {
		showlocation(w, r, sqllocation, loc.NumBoxes)
	}

}

func showlocation(w http.ResponseWriter, r *http.Request, sqllocation string, NumBoxes int) {

	// Header for box listing by location
	var locboxtablehdr = `
<table class="boxlist">
<thead>
<tr>
<th class="boxid"><a title="&#8645;" href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}&` + Param_Labels["order"] + `=boxid{{if .Desc}}&` + Param_Labels["desc"] + `=boxid{{end}}">` + prefs.Field_Labels["boxid"] + `</a></th>
<th class="storeref"><a title="&#8645;" href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}&` + Param_Labels["order"] + `=storeref{{if .Desc}}&` + Param_Labels["desc"] + `=storeref{{end}}">` + prefs.Field_Labels["storeref"] + `</a></th>
<th class="contents"><a title="&#8645;" href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}&` + Param_Labels["order"] + `=overview{{if .Desc}}&` + Param_Labels["desc"] + `=overview{{end}}">` + prefs.Field_Labels["overview"] + `</a></th>
<th class="boxid"><a title="&#8645;" href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}&` + Param_Labels["order"] + `=numdocs{{if .Desc}}&` + Param_Labels["desc"] + `=numdocs{{end}}">` + prefs.Field_Labels["numdocs"] + `</a></th>
<th class="boxid"><a title="&#8645;" href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}&` + Param_Labels["order"] + `=min_review_date{{if .Desc}}&` + Param_Labels["desc"] + `=min_review_date{{end}}">` + prefs.Field_Labels["review_date"] + `</a></th>
</tr>
</thead>
<tbody>
`

	if r.FormValue(Param_Labels["location"]) == "" {
		show_search(w, r)
		return
	}

	var loc locationlistvars
	loc.Single = true
	loc.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])
	loc.Location, _ = url.QueryUnescape(r.FormValue(Param_Labels["location"]))
	loc.LocationUrl = r.FormValue(Param_Labels["location"])
	loc.NumBoxes, _ = strconv.Atoi(getValueFromDB("SELECT Count(*) As rex FROM boxes WHERE location='"+sqllocation+"'", "rex", "0"))
	loc.NumBoxesX = commas(loc.NumBoxes)
	loc.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"]) || r.FormValue(Param_Labels["order"]) == ""

	temp, err := template.New("locboxtablehdr").Parse(locboxtablehdr)
	if err != nil {
		panic(err)
	}
	err = temp.Execute(w, loc)
	if err != nil {
		panic(err)
	}
	sqlx := "SELECT storeref,boxid,location,overview,numdocs,min_review_date,max_review_date FROM boxes WHERE location='" + sqllocation + "'"

	if r.FormValue(Param_Labels["order"]) != "" {
		sqlx += " ORDER BY " + r.FormValue(Param_Labels["order"])
		if r.FormValue(Param_Labels["desc"]) != "" {
			sqlx += " DESC"
		}
	} else {
		sqlx += " ORDER BY boxid"
	}

	//sqllimit := emit_page_anchors(w, r, "locations?"+Param_Labels["location"]+"="+url.QueryEscape(r.FormValue(Param_Labels["location"])), loc.NumBoxes)
	sqllimit := emit_page_anchors(w, r, "locations", loc.NumBoxes)
	//fmt.Print("DEBUG: " + sqlx)
	rows, err := DBH.Query(sqlx + sqllimit)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var bv boxvars
	bv.Single = r.FormValue(Param_Labels["location"]) != ""
	bv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"]) || r.FormValue(Param_Labels["order"]) == ""

	temp, err = template.New("locboxtablerow").Parse(locboxtablerow)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var mindate, maxdate string
		rows.Scan(&bv.Storeref, &bv.Boxid, &bv.Location, &bv.Contents, &bv.NumFiles, &mindate, &maxdate)
		if mindate == maxdate {
			bv.Date = mindate
			bv.Single = true
		} else {
			bv.Date = mindate + " to " + maxdate
			bv.Single = false
		}
		bv.LocationUrl = template.URLQueryEscaper(loc.Location)
		bv.StorerefUrl = template.URLQueryEscaper(bv.Storeref)
		bv.BoxidUrl = template.URLQueryEscaper(bv.Boxid)
		err = temp.Execute(w, bv)
		if err != nil {
			panic(err)
		}
	}
	//	showBoxfiles(w, r, sqlboxid)

}

func showlocationfiles(w http.ResponseWriter, r *http.Request, boxid string) {

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

	temp, err := template.New("boxfileshdr").Parse(boxfileshdr)
	if err != nil {
		panic(err)
	}

	var bfv boxfilevars
	bfv.Boxid = boxid
	bfv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])

	err = temp.Execute(w, bfv)
	if err != nil {
		panic(err)
	}

	nrows := 0
	temp, err = template.New("boxfilesline").Parse(boxfilesline)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		rows.Scan(&bfv.Owner, &bfv.Client, &bfv.Name, &bfv.Contents, &bfv.Date)
		bfv.OwnerUrl = template.URLQueryEscaper(bfv.Owner)
		bfv.ClientUrl = template.URLQueryEscaper(bfv.Client)
		err = temp.Execute(w, bfv)
		if err != nil {
			panic(err)
		}

		nrows++
	}
	temp, err = template.New("boxfilestrailer").Parse(boxfilestrailer)
	temp.Execute(w, "")
	if err != nil {
		panic(err)
	}

}
