package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func exec_search(w http.ResponseWriter, r *http.Request) {

	// This needs to be here in order to collect runtime values from prefs

	var searchResultsHdr1 = `
<p>{{if .Find}}I was looking for <span class="searchedfor">{{.Find}}{{if .Field}} in {{.Field}}{{end}}{{if .Locations}} [` + prefs.Field_Labels["location"] + `: {{.Locations}}]{{end}}{{if .Owners}} [` + prefs.Field_Labels["owner"] + `: {{.Owners}}]{{end}}</span> and{{end}} I found {{if .Found0}}nothing, nada, rien, zilch.{{end}}{{if .Found1}}just the one match.{{end}}{{if .Found2}}{{.Found}} matches.{{end}}</p>
`

	var searchResultsHdr2 = `
	<table class="searchresults">
	<thead>
	<tr>
	<th class="ourbox"><a href="/find?` + Param_Labels["find"] + `={{.FindUrl}}&` + Param_Labels["order"] + `=boxid{{if .Desc}}&` + Param_Labels["desc"] + `=boxid{{end}}">` + prefs.Field_Labels["boxid"] + `</a></th>
	<th class="owner"><a href="/find?` + Param_Labels["find"] + `={{.FindUrl}}&` + Param_Labels["order"] + `=owner{{if .Desc}}&` + Param_Labels["desc"] + `=owner{{end}}">` + prefs.Field_Labels["owner"] + `</a></th>
	<th class="client"><a href="/find?` + Param_Labels["find"] + `={{.FindUrl}}&` + Param_Labels["order"] + `=client{{if .Desc}}&` + Param_Labels["desc"] + `=client{{end}}">` + prefs.Field_Labels["client"] + `</a></th>
	<th class="name"><a href="/find?` + Param_Labels["find"] + `={{.FindUrl}}&` + Param_Labels["order"] + `=name{{if .Desc}}&` + Param_Labels["desc"] + `=name{{end}}">` + prefs.Field_Labels["name"] + `</a></th>
	<th class="contents"><a href="/find?` + Param_Labels["find"] + `={{.FindUrl}}&` + Param_Labels["order"] + `=contents{{if .Desc}}&` + Param_Labels["desc"] + `=contents{{end}}">` + prefs.Field_Labels["contents"] + `</a></th>
	<th class="date"><a href="/find?` + Param_Labels["find"] + `={{.FindUrl}}&` + Param_Labels["order"] + `=review_date{{if .Desc}}&` + Param_Labels["desc"] + `=review_date{{end}}">` + prefs.Field_Labels["review_date"] + `</a></th>
	</tr>
	</thead>
	<tbody>
	`

	start_html(w, r)

	var sqlx = ` FROM contents LEFT JOIN boxes ON contents.boxid=boxes.boxid `
	var wherex = ``
	if r.FormValue(Param_Labels["find"]) != "" {
		if r.FormValue(Param_Labels["field"]) != "" {

			if r.FormValue(Param_Labels["field"]) == "review_date" {
				wherex += `review_date LIKE '?%'`
			} else if r.FormValue(Param_Labels["field"]) == "contents" {
				wherex += `contents LIKE '%?%'`
			} else if r.FormValue(Param_Labels["field"]) == "name" {
				wherex += `name LIKE '%?%'`
			} else {
				wherex += r.FormValue(Param_Labels["field"]) + `= '?'`
			}
		} else {
			wherex += `(
				(contents.boxid = '?')
			OR 	(boxes.storeref = '?') 
			OR 	(boxes.overview LIKE '%?%')
        	OR 	(contents.owner = '?') 
        	OR 	(contents.client = '?') 
        	OR 	(contents.contents LIKE '%?%') 
        	OR 	(contents.name LIKE '%?%')
			OR 	(contents.review_date = '?')
			OR 	(contents.review_date LIKE '?%')
			)
		`
		}
	}

	session, err := store.Get(r, cookie_name)
	checkerr(err)

	if session.Values["locations"] != nil {
		if wherex != "" {
			wherex += " AND "
		}
		wherex += " location In (" + session.Values["locations"].(string) + ")"
	}
	if session.Values["owners"] != nil {
		if wherex != "" {
			wherex += " AND "
		}
		wherex += " owner In (" + session.Values["owners"].(string) + ")"
	}
	x, _ := url.QueryUnescape(r.FormValue(Param_Labels["find"]))
	wherex = strings.ReplaceAll(wherex, "?", strings.ReplaceAll(x, "'", "''"))
	if wherex != "" {
		sqlx += " WHERE " + wherex
	}
	if r.FormValue(Param_Labels["order"]) != "" {
		sqlx += " ORDER BY Upper(Trim(contents." + r.FormValue(Param_Labels["order"]) + "))"
		if r.FormValue(Param_Labels["desc"]) != "" {
			sqlx += " DESC"
		}
	}

	//fmt.Println("DEBUG: " + sqlx)
	FoundRecCount, _ := strconv.Atoi(getValueFromDB("SELECT Count(*) AS Rexx"+sqlx, "Rexx", "0"))

	var res searchResultsVar

	if session.Values["locations"] != nil {
		res.Locations = session.Values["locations"].(string)
	}
	if session.Values["owners"] != nil {
		res.Owners = session.Values["owners"].(string)
	}
	res.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])

	res.Boxid = order_dir(r, "boxid")
	res.BoxidUrl = template.URLQueryEscaper(res.Boxid)
	res.Owner = order_dir(r, "owner")
	res.OwnerUrl = template.URLQueryEscaper(res.Owner)
	res.Client = order_dir(r, "client")
	res.ClientUrl = template.URLQueryEscaper(res.Client)
	res.Name = order_dir(r, "name")
	res.Date = order_dir(r, "review_date")
	res.Find = x
	res.FindUrl = template.URLQueryEscaper(res.Find)
	res.Found = commas(FoundRecCount)
	res.Found0 = FoundRecCount == 0
	res.Found1 = FoundRecCount == 1
	res.Found2 = FoundRecCount > 1
	res.Field = prefs.Field_Labels[r.FormValue(Param_Labels["field"])]

	html, err := template.New("searchResultsHdr1").Parse(searchResultsHdr1)
	checkerr(err)
	html.Execute(w, res)

	flds := "contents.boxid,contents.owner,contents.client,contents.name,contents.contents,contents.review_date,boxes.storeref,boxes.overview"

	sqllimit := emit_page_anchors(w, r, "find", FoundRecCount)

	//fmt.Printf("DEBUG: sql = SELECT %v%v%v\n", flds, sqlx, sqllimit)

	rows, err := DBH.Query("SELECT " + flds + sqlx + sqllimit)
	if err != nil {
		fmt.Printf("Omg! %v\n", sqlx)
		panic(err)
	}
	html, err = template.New("searchResultsHdr2").Parse(searchResultsHdr2)
	checkerr(err)
	html.Execute(w, res)

	html, err = template.New("searchResultsLine").Parse(searchResultsLine)
	checkerr(err)
	for rows.Next() {
		rows.Scan(&res.Boxid, &res.Owner, &res.Client, &res.Name, &res.Contents, &res.Date, &res.Storeref, &res.Overview)
		res.BoxidUrl = template.URLQueryEscaper(res.Boxid)
		res.OwnerUrl = template.URLQueryEscaper(res.Owner)
		res.ClientUrl = template.URLQueryEscaper(res.Client)
		res.StorerefUrl = template.URLQueryEscaper(res.Storeref)
		err = html.Execute(w, res)
		if err != nil {
			panic(err)
		}
	}
	html, _ = template.New("searchResultsTrailer").Parse(searchResultsTrailer)
	html.Execute(w, "")

}

func show_search_params(w http.ResponseWriter, r *http.Request) {

	var paramsHTML = `

	<form action="/params">
	<main>
	<p>The settings you choose here will be used to restrict searches until you reset them or until your session ends.</p>
	<p><input data-triggered="0" id="savesettings" class="btn" name="` + Param_Labels["savesettings"] + `" disabled onclick="this.setAttribute('data-triggered','1');return true;" type="submit" value="Save settings"></p>`

	var params struct {
		Lrange    string
		Locations string

		Orange string
		Owners string
	}

	var locnsRadios = `
	<p>
	<input type="radio" id="range_all" name="l` + Param_Labels["range"] + `" value="` + Param_Labels["all"] + `" {{if eq .Lrange "` + Param_Labels["all"] + `"}}checked{{end}} onclick="param_select_locations(this.checked);">
	<label for="range_all"> All </label> &nbsp;&nbsp;&nbsp; 
	<input type="radio" id="range_sel" name="l` + Param_Labels["range"] + `" value="` + Param_Labels["selected"] + `" {{if ne .Lrange "` + Param_Labels["all"] + `"}}checked{{end}} onclick="param_select_locations(!this.checked);">
	<label for="range_sel"> Selected only </label>&nbsp;&nbsp;&nbsp;
	</p>
`

	var ownerRadios = `
	<p>
	<input type="radio" id="orange_all" name="o` + Param_Labels["range"] + `" value="` + Param_Labels["all"] + `" {{if eq .Orange "` + Param_Labels["all"] + `"}}checked{{end}} onclick="param_select_owners(this.checked);">
	<label for="orange_all"> All </label> &nbsp;&nbsp;&nbsp; 
	<input type="radio" id="orange_sel" name="o` + Param_Labels["range"] + `" value="` + Param_Labels["selected"] + `" {{if ne .Orange "` + Param_Labels["all"] + `"}}checked{{end}} onclick="param_select_owners(!this.checked);">
	<label for="orange_sel"> Selected only </label>&nbsp;&nbsp;&nbsp;
	</p>
`

	r.ParseForm()
	session, err := store.Get(r, cookie_name)
	checkerr(err)

	if session.Values["locations"] == nil {
		params.Lrange = Param_Labels["all"]
		params.Locations = ""
	} else {
		params.Lrange = Param_Labels["selected"]
		params.Locations = session.Values["locations"].(string)
	}
	if session.Values["owners"] == nil {
		params.Orange = Param_Labels["all"]
		params.Owners = ""
	} else {
		params.Orange = Param_Labels["selected"]
		params.Owners = session.Values["owners"].(string)
	}

	// Update settings

	if r.FormValue("l"+Param_Labels["range"]) != "" {
		params.Lrange = r.FormValue("l" + Param_Labels["range"])
		//fmt.Printf("DEBUG: params lrange is %v\n", params.Lrange)
		var locs []string
		for _, x := range r.Form[Param_Labels["location"]] {
			locs = append(locs, "'"+strings.ReplaceAll(x, "'", "''")+"'")
		}
		params.Locations = strings.Join(locs, ",")
		if params.Lrange == Param_Labels["all"] {
			session.Values["locations"] = nil
		} else {
			session.Values["locations"] = params.Locations
		}
		err = store.Save(r, w, session)
		checkerr(err)
	}
	if r.FormValue("o"+Param_Labels["range"]) != "" {
		params.Orange = r.FormValue("o" + Param_Labels["range"])
		//fmt.Printf("DEBUG: params orange is %v\n", params.Orange)
		var owners []string
		for _, x := range r.Form[Param_Labels["owner"]] {
			owners = append(owners, "'"+strings.ReplaceAll(x, "'", "''")+"'")
		}
		params.Owners = strings.Join(owners, ",")
		if params.Orange == Param_Labels["all"] {
			session.Values["owners"] = nil
		} else {
			session.Values["owners"] = params.Owners
		}
		err = store.Save(r, w, session)
		checkerr(err)
	}

	if r.FormValue(Param_Labels["savesettings"]) != "" {
		show_search(w, r)
		return
	}
	start_html(w, r)

	//fmt.Fprintf(w, `DEBUG: %v<hr>`, r)
	//fmt.Fprintf(w, `DEBUG: %v<hr>`, r.Form["qlo"])
	//fmt.Fprintf(w, `DEBUG: %v<hr>`, strings.Join(r.Form["qlo"], ","))

	fmt.Fprintln(w, paramsHTML)

	fmt.Fprintln(w, `<div id="locationfilter">`)
	fmt.Fprintln(w, `<h2>`+prefs.Field_Labels["location"]+`s</h2>`)

	temp, err := template.New("locnsRadios").Parse(locnsRadios)
	checkerr(err)
	temp.Execute(w, params)

	hideshow := ""
	if params.Lrange == Param_Labels["all"] {
		hideshow = " hide "
	}
	sqlx := "SELECT location FROM locations ORDER BY location"
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	fmt.Fprintf(w, `<div class="filteritems%v">`, hideshow)
	n := 0
	nmax := prefs.DefaultPagesize
	for rows.Next() {
		n++
		if n > nmax {
			fmt.Fprintf(w, `</div><div class="filteritems%v">`, hideshow)
			n = 1
		}
		var locn string
		rows.Scan(&locn)
		checked := ""
		if strings.Contains(params.Locations, "'"+strings.ReplaceAll(locn, "'", "''")+"'") {
			checked = " checked "
		}
		fmt.Fprintf(w, `<input id="cb_%v" type="checkbox" onclick="enableSaveSettings();" name="`+Param_Labels["location"]+`" value="%v" %v> `, locn, locn, checked)
		fmt.Fprintf(w, ` <label for="cb_%v">%v</label><br>`, locn, locn)
	}
	fmt.Fprintln(w, "</div></div>")

	fmt.Fprintln(w, `<div id="ownerfilter">`)
	fmt.Fprintln(w, `<h2>`+prefs.Field_Labels["owner"]+`s</h2>`)
	temp, err = template.New("ownerRadios").Parse(ownerRadios)
	checkerr(err)
	temp.Execute(w, params)
	sqlx = "SELECT DISTINCT Trim(owner) As ownerx FROM contents ORDER BY ownerx"
	rows, err = DBH.Query(sqlx)
	checkerr(err)
	hideshow = ""
	if params.Orange == Param_Labels["all"] {
		hideshow = " hide "
	}
	//fmt.Printf("Orange is %v; hideshow is %v\n", params.Orange, hideshow)
	fmt.Fprintf(w, `<div class="filteritems%v">`, hideshow)
	n = 0
	nmax = prefs.DefaultPagesize
	for rows.Next() {
		n++
		if n > nmax {
			fmt.Fprintf(w, `</div><div class="filteritems%v">`, hideshow)
			n = 1
		}
		var locn string
		rows.Scan(&locn)
		checked := ""
		if strings.Contains(params.Owners, "'"+strings.ReplaceAll(locn, "'", "''")+"'") {
			checked = " checked "
		}
		fmt.Fprintf(w, `<input id="cb_%v" type="checkbox" name="`+Param_Labels["owner"]+`" onclick="enableSaveSettings();" value="%v" %v> `, locn, locn, checked)
		fmt.Fprintf(w, ` <label for="cb_%v">%v</label><br>`, locn, locn)
	}

	fmt.Fprintln(w, "</div></div>")

}

func show_search(w http.ResponseWriter, r *http.Request) {

	var sv searchVars

	var searchHTML = `
<p>I'm currently minding <strong>{{.NumDocsX}}</strong> individual files packed into
<strong>{{.NumBoxesX}}</strong> boxes stored in <strong>{{.NumLocnsX}}
</strong> locations.</p>

<form action="/find" onsubmit="if (this.fld.value=='') { this.fld.name=''; }">
<main>You can search the archives using a simple textsearch by entering the text you're looking for
here <input type="text" autofocus name="` + Param_Labels["find"] + `"/>
<details title="Fields to search" style="display:inline;">
<summary><strong>&#8799;</strong></summary>
<select id="fld" name="` + Param_Labels["field"] + `">
<option value="">any field</option>
<option value="client">` + prefs.Field_Labels["client"] + `</option>
<option value="name">` + prefs.Field_Labels["name"] + `</option>
<option value="owner">` + prefs.Field_Labels["owner"] + `</option>
<option value="contents">` + prefs.Field_Labels["contents"] + `</option>
<option value="review_date">` + prefs.Field_Labels["review_date"] + `</option>
<option value="storeref">` + prefs.Field_Labels["storeref"] + `</option>
<option value="boxid">` + prefs.Field_Labels["boxid"] + `</option>
</select>
</details>
<input type="submit" value="Find it!"/><br />
You can enter a partner's initials, a client number or name, a common term such as <em>tax</em> or a review date or year.<br>
Just enter the terms you're looking for, no quote marks, ANDs, ORs, etc.</main></form>
<p>If you want to search only for records belonging to particular ` + prefs.Field_Labels["owner"] + `s or ` + prefs.Field_Labels["location"] + `s, <a href="/params">specify search options here</a>.</p>
<p>{{if or .Locations .Owners}}Current search restrictions:- {{if .Locations}}<strong>` + prefs.Field_Labels["location"] + `: {{.Locations}};</strong> {{end}} {{if .Owners}}<strong>` + prefs.Field_Labels["owner"] + `: {{.Owners}};</strong> {{end}} {{end}}</p>
`

	start_html(w, r)

	session, err := store.Get(r, cookie_name)
	checkerr(err)

	if session.Values["locations"] != nil {
		sv.Locations = session.Values["locations"].(string)
	}
	if session.Values["owners"] != nil {
		sv.Owners = session.Values["owners"].(string)
	}
	sv.Apptitle = prefs.AppTitle
	sv.NumBoxes, _ = strconv.Atoi(getValueFromDB("SELECT Count(*) As Rex FROM boxes", "Rex", "-1"))
	sv.NumBoxesX = commas(sv.NumBoxes)
	sv.NumDocs, _ = strconv.Atoi(getValueFromDB("SELECT Count(*) As Rex FROM contents", "Rex", "-1"))
	sv.NumDocsX = commas(sv.NumDocs)
	sv.NumLocns, _ = strconv.Atoi(getValueFromDB("SELECT Count(*) As Rex FROM locations", "Rex", "-1"))
	sv.NumLocnsX = commas(sv.NumLocns)

	html, err := template.New("searchHTML").Parse(searchHTML)
	checkerr(err)

	html.Execute(w, sv)

	fmt.Fprintln(w, "</body></html>")
}
