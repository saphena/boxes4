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
	if r.FormValue(Param_Labels["find"]) != "" {
		if r.FormValue(Param_Labels["field"]) != "" {
			sqlx += `WHERE `
			if r.FormValue((Param_Labels["field"])) == "review_date" {
				sqlx += `review_date LIKE '?%'`
			} else if r.FormValue((Param_Labels["field"])) == "contents" {
				sqlx += `contents LIKE '%?%'`
			} else if r.FormValue((Param_Labels["field"])) == "name" {
				sqlx += `name LIKE '%?%'`
			} else {
				sqlx += r.FormValue((Param_Labels["field"])) + `= '?'`
			}
		} else {
			sqlx += `WHERE ((contents.boxid = '?')
			OR (boxes.storeref = '?') 
			OR (boxes.overview LIKE '%?%')
        	OR (contents.owner = '?') 
        	OR (contents.client = '?') 
        	OR (contents.contents LIKE '%?%') 
        	OR (contents.name LIKE '%?%')) 
			OR (contents.review_date = '?')
			OR (contents.review_date LIKE '?%')
		`
		}
	}
	x, _ := url.QueryUnescape(r.FormValue(Param_Labels["find"]))
	sqlx = strings.ReplaceAll(sqlx, "?", strings.ReplaceAll(x, "'", "''"))
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

func show_search(w http.ResponseWriter, r *http.Request) {

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
Just enter the words you're looking for, no quote marks, ANDs, ORs, etc.</main></form>
<p>If you want to search only for records belonging to particular ` + prefs.Field_Labels["owner"] + `s or ` + prefs.Field_Labels["location"] + `s, <a href="index.php?CMD=PARAMS">specify search options here</a>.</p>
<form action="/boxes"
    onsubmit="return !isBadLength(this.` + Param_Labels["boxid"] + `,1,
    'I\'m sorry, computers don\'t do guessing; you have to tell me which box to show you.\n\nPerhaps you want to see a list of boxes available in which case you should click on [boxes] above.');">
<p>If you want to look at a particular box, enter its ID here
<input type="text" name="` + Param_Labels["boxid"] + `" size="10"/><input type="submit" value="Show box"/></p></form>
`

	start_html(w, r)

	searchVars.Apptitle = "DOCUMENT ARCHIVES"
	searchVars.NumBoxes, _ = strconv.Atoi(getValueFromDB("SELECT Count(*) As Rex FROM boxes", "Rex", "-1"))
	searchVars.NumBoxesX = commas(searchVars.NumBoxes)
	searchVars.NumDocs, _ = strconv.Atoi(getValueFromDB("SELECT Count(*) As Rex FROM contents", "Rex", "-1"))
	searchVars.NumDocsX = commas(searchVars.NumDocs)
	searchVars.NumLocns, _ = strconv.Atoi(getValueFromDB("SELECT Count(*) As Rex FROM locations", "Rex", "-1"))
	searchVars.NumLocnsX = commas(searchVars.NumLocns)

	html, err := template.New("searchHTML").Parse(searchHTML)
	checkerr(err)

	html.Execute(w, searchVars)

	fmt.Fprintln(w, "</body></html>")
}
