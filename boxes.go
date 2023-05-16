package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func showboxes(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	if r.FormValue(Param_Labels["chgboxlocn"]) != "" {
		ajax_change_box_location(w, r)
		return
	}

	if r.FormValue(Param_Labels["savecontent"]) != "" {
		ajax_update_content_line(w, r)
		return
	}

	if r.FormValue(Param_Labels["delcontent"]) != "" {
		ajax_delete_content_line(w, r)
		return
	}

	if r.FormValue(Param_Labels["newcontent"]) != "" {
		ajax_add_new_content(w, r)
		return
	}

	if r.FormValue(Param_Labels["client"]) != "" {
		ajax_fetch_name_list(w, r)
		return
	}
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
	checkerr(err)
	defer rows.Close()

	var box boxvars
	box.Single = r.FormValue(Param_Labels["boxid"]) != ""
	box.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"]) || r.FormValue(Param_Labels["order"]) == ""

	html, err := template.New("").Parse(boxtablehdr)
	checkerr(err)
	err = html.Execute(w, box)
	checkerr(err)

	html, err = template.New("").Parse(boxtablerow)
	checkerr(err)
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
	<input type="hidden" id="AutosaveSeconds" value="` + strconv.Itoa(prefs.AutosaveSeconds) + `">
<table class="boxheader">


<tr><td class="vlabel">{{if .Single}}{{else}}<a title="&#8645;" href="/boxes?` + Param_Labels["boxid"] + `={{.BoxidUrl}}&` + Param_Labels["order"] + `=boxid&` + Param_Labels["desc"] + `=boxid">{{end}}` + prefs.Field_Labels["boxid"] + `{{if .Single}}{{else}}</a>{{end}} : </td><td id="boxboxid" class="vdata">{{.Boxid}}</td></tr>
<tr>
<td class="vlabel">` + prefs.Field_Labels["location"] + ` : </td>
<td class="vdata">{{if .UpdateOK}}#LOCSELECTOR#{{else}}<a href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}">{{.Location}}</a>{{end}}</td>
</tr>
<tr><td class="vlabel">` + prefs.Field_Labels["storeref"] + ` : </td><td class="vdata"><a href="/find?` + Param_Labels["find"] + `={{.StorerefUrl}}&` + Param_Labels["field"] + `=storeref">{{.Storeref}}</a></td></tr>
<tr><td class="vlabel">` + prefs.Field_Labels["contents"] + ` : </td><td class="vdata">{{.Contents}}</td></tr>
<tr><td class="vlabel">` + prefs.Field_Labels["numdocs"] + ` : </td><td id="boxnumfiles" class="vdata numdocs">{{.NumFilesX}}</td></tr>
<tr><td class="vlabel">` + prefs.Field_Labels["review_date"] + ` : </td><td id="boxdates" class="vdata center">{{.Date}}</td></tr>

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
	checkerr(err)
	defer rows.Close()
	var bv boxvars
	bv.Single = r.FormValue(Param_Labels["boxid"]) != ""
	bv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])
	bv.UpdateOK = runvars.Updating
	bv.DeleteOK = runvars.Updating

	if !rows.Next() {
		fmt.Fprintf(w, "<p>Bugger! %v</p>", r.FormValue(Param_Labels["boxid"]))
		return
	}
	//xx := prefs.Field_Labels["boxid"]
	//fmt.Printf("xx=%v\n", xx)
	var mindate, maxdate string
	rows.Scan(&bv.Storeref, &bv.Boxid, &bv.Location, &bv.Contents, &bv.NumFiles, &mindate, &maxdate)
	if mindate == maxdate {
		bv.Date = mindate
	} else {
		bv.Date = mindate + " to " + maxdate
	}
	bv.NumFilesX = commas(bv.NumFiles)
	t := strings.ReplaceAll(boxhtml, "#LOCSELECTOR#", generateLocationPicklist(bv.Location, "change_box_location(this);"))
	html, err := template.New("main").Parse(t)
	checkerr(err)
	err = html.Execute(w, bv)
	checkerr(err)
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
	sqlx := "SELECT owner,client,name,contents,review_date,id FROM contents WHERE boxid='" + boxid + "'"

	if r.FormValue(Param_Labels["order"]) != "" {
		sqlx += " ORDER BY TRIM(contents." + r.FormValue(Param_Labels["order"]) + ")"
		if r.FormValue(Param_Labels["desc"]) != "" {
			sqlx += " DESC"
		}
	} else {
		sqlx += " ORDER BY id"
	}
	rows, _ := DBH.Query(sqlx + sqllimit)
	defer rows.Close()

	html, err := template.New("").Parse(boxfileshdr)
	checkerr(err)

	var bfv boxfilevars
	bfv.Boxid = boxid
	bfv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])
	bfv.DeleteOK = runvars.Updating
	bfv.UpdateOK = runvars.Updating

	err = html.Execute(w, bfv)
	checkerr(err)

	if runvars.Updating {
		//fmt.Fprint(w, newboxcontentline)
		temp, err := template.New("newboxcontentline").Parse(newboxcontentline)
		checkerr(err)
		err = temp.Execute(w, bfv)
	}
	nrows := 0

	for rows.Next() {
		rows.Scan(&bfv.Owner, &bfv.Client, &bfv.Name, &bfv.Contents, &bfv.Date, &bfv.Id)
		bfv.OwnerUrl = template.URLQueryEscaper(bfv.Owner)
		bfv.ClientUrl = template.URLQueryEscaper(bfv.Client)

		t := strings.ReplaceAll(boxfilesline, "#DATESELECTORS#", generateDatePicklist(bfv.Date, Param_Labels["review_date"], "contentSaveNeeded(this.parentElement);"))
		html, err = template.New("").Parse(t)
		checkerr(err)

		err = html.Execute(w, bfv)
		if err != nil {
			panic(err)
		}

		nrows++
	}
	html, err = template.New("").Parse(boxfilestrailer)
	html.Execute(w, "")
	checkerr(err)

	emit_owner_list(w)
	emit_client_list(w)
	emit_name_list(w)

}

func ajax_fetch_name_list(w http.ResponseWriter, r *http.Request) {

	client := r.FormValue(Param_Labels["client"])
	sqlx := "SELECT DISTINCT Trim(name) FROM contents"
	if client != "" {
		sqlx += " WHERE client='" + strings.ReplaceAll(client, "'", "''") + "'"
	}
	sqlx += " ORDER BY Trim(name)"
	fmt.Println("DEBUG: " + sqlx)
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	fmt.Fprint(w, `{"res":"ok","names":[`)
	emitComma := false
	for rows.Next() {
		var name string
		rows.Scan(&name)
		if emitComma {
			fmt.Fprint(w, ",")
		}
		fmt.Fprintf(w, `"%v"`, name)
		emitComma = true
	}
	fmt.Fprint(w, `]}`)

}

func ajax_add_new_content(w http.ResponseWriter, r *http.Request) {

	boxid := r.FormValue(Param_Labels["newcontent"])
	owner := r.FormValue(Param_Labels["owner"])
	client := r.FormValue(Param_Labels["client"])
	name := r.FormValue(Param_Labels["name"])
	contents := r.FormValue(Param_Labels["contents"])
	review := "2030-01-01"

	// Let's apply some lazy help

	re := regexp.MustCompile(`.*[A-Z]`) // Check for at least one uppercase letter
	if contains(prefs.FixLazyTyping, "name") && !re.MatchString(name) {
		name = fixAllLowercase(name)
	}
	if contains(prefs.FixLazyTyping, "contents") && !re.MatchString(contents) {
		contents = fixAllLowercase(contents)
	}

	sqlx := "INSERT INTO contents (boxid,review_date,contents,owner,name,client) VALUES("
	sqlx += "'" + safesql(boxid) + "'"
	sqlx += ",'" + review + "'"
	sqlx += ",'" + safesql(contents) + "'"
	sqlx += ",'" + safesql(owner) + "'"
	sqlx += ",'" + safesql(name) + "'"
	sqlx += ",'" + safesql(client) + "'"
	sqlx += ")"

	fmt.Println("DEBUG: " + sqlx)
	res := DBExec(sqlx)
	n, err := res.RowsAffected()
	checkerr(err)
	if n < 1 {
		fmt.Fprint(w, `{"res":"Insertion failed!"}`)
		return
	}
	nf, ld, hd := update_ajax_box_contents(boxid)
	fmt.Fprintf(w, `{"res":"ok","nfiles":"%v","lodate":"%v","hidate":"%v"}`, nf, ld, hd)

}

func ajax_delete_content_line(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue(Param_Labels["delcontent"])
	owner := r.FormValue(Param_Labels["owner"])
	client := r.FormValue(Param_Labels["client"])
	boxid := r.FormValue(Param_Labels["boxid"])

	sqlx := "DELETE FROM contents WHERE id=" + id
	sqlx += " AND owner='" + safesql(owner) + "'"
	sqlx += " AND client='" + safesql(client) + "'"
	fmt.Println("DEBUG: " + sqlx)
	res := DBExec(sqlx)
	n, err := res.RowsAffected()
	checkerr(err)
	if n < 1 {
		fmt.Fprint(w, `{"res":"Deletion failed!"}`)
		return
	}

	nf, ld, hd := update_ajax_box_contents(boxid)
	fmt.Fprintf(w, `{"res":"ok","nfiles":"%v","lodate":"%v","hidate":"%v"}`, nf, ld, hd)

}

func ajax_update_content_line(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue(Param_Labels["savecontent"])
	owner := r.FormValue(Param_Labels["owner"])
	client := r.FormValue(Param_Labels["client"])
	name := r.FormValue(Param_Labels["name"])
	contents := r.FormValue(Param_Labels["contents"])
	review := r.FormValue(Param_Labels["review_date"])
	boxid := r.FormValue(Param_Labels["boxid"])

	sqlx := "UPDATE contents SET "
	sqlx += " owner='" + safesql(owner) + "'"
	sqlx += ",client='" + safesql(client) + "'"
	sqlx += ",name='" + safesql(name) + "'"
	sqlx += ",contents='" + safesql(contents) + "'"
	sqlx += ",review_date='" + safesql(review) + "'"

	sqlx += " WHERE id=" + id

	fmt.Println("DEBUG: " + sqlx)
	res := DBExec(sqlx)
	n, err := res.RowsAffected()
	checkerr(err)
	if n < 1 {
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}

	nf, ld, hd := update_ajax_box_contents(boxid)
	fmt.Fprintf(w, `{"res":"ok","nfiles":"%v","lodate":"%v","hidate":"%v"}`, nf, ld, hd)
}

func update_ajax_box_contents(boxid string) (int, string, string) {

	sqlx := "SELECT review_date FROM contents WHERE boxid='" + safesql(boxid) + "'"
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	nfiles := 0
	lodate := "9999-12-31"
	hidate := "0000-01-01"
	for rows.Next() {
		var dt string
		rows.Scan(&dt)
		nfiles++
		if dt < lodate {
			lodate = dt
		}
		if dt > hidate {
			hidate = dt
		}
	}
	rows.Close()
	sqlx = "UPDATE boxes SET numdocs=" + strconv.Itoa(nfiles)
	sqlx += ",min_review_date='" + lodate + "'"
	sqlx += ",max_review_date='" + hidate + "'"
	sqlx += "WHERE boxid='" + safesql(boxid) + "'"
	fmt.Println("DEBUG: " + sqlx)
	DBExec(sqlx)

	return nfiles, lodate, hidate

}

func ajax_change_box_location(w http.ResponseWriter, r *http.Request) {

	boxid := r.FormValue(Param_Labels["boxid"])
	locn := r.FormValue(Param_Labels["chgboxlocn"])

	sqlx := "SELECT location FROM locations WHERE location='" + safesql(locn) + "'"
	if getValueFromDB(sqlx, "location", "") == "" {
		fmt.Fprint(w, `{"res":"`+prefs.Field_Labels["location"]+` doesn't exist"}`)
		return
	}
	sqlx = "UPDATE boxes SET location='" + safesql(locn) + "' WHERE boxid='" + safesql(boxid) + "'"
	fmt.Println("DEBUG: " + sqlx)
	res := DBExec(sqlx)
	n, err := res.RowsAffected()
	checkerr(err)
	if n < 1 {
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}

	fmt.Fprint(w, `{"res":"ok"}`)
}
