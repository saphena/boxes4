package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func ajax_add_location(w http.ResponseWriter, r *http.Request) {
	// form already parsed so just get on with it

	newloc := r.FormValue(Param_Labels["newloc"])
	newlocsql := strings.ReplaceAll(newloc, "'", "''")
	sqlx := "SELECT Count(location) As rex FROM locations WHERE location LIKE '" + newlocsql + "'"
	dupe := getValueFromDB(sqlx, "0") != "0"
	if dupe {
		fmt.Fprint(w, `{"res":"That `+prefs.Field_Labels["location"]+` already exists!"}`)
		return
	}
	sqlx = "INSERT INTO locations (location) VALUES('" + newlocsql + "')"
	res := DBExec(sqlx)
	n, err := res.RowsAffected()
	checkerr(err)
	if n < 1 {
		fmt.Fprint(w, `{"res":"Insert failed!"}`)
		return
	}

	fmt.Fprint(w, `{"res":"ok"}`)

}

func ajax_del_location(w http.ResponseWriter, r *http.Request) {
	// form already parsed so just get on with it

	oldloc := r.FormValue(Param_Labels["delloc"])
	oldlocsql := strings.ReplaceAll(oldloc, "'", "''")
	sqlx := "SELECT Count(boxid) As rex FROM boxes WHERE location LIKE '" + oldlocsql + "'"
	dupe := getValueFromDB(sqlx, "0")
	if dupe != "0" {
		fmt.Fprint(w, `{"res":"That `+prefs.Field_Labels["location"]+` contains at least one box!"}`)
		return
	}
	sqlx = "DELETE FROM locations WHERE location='" + oldlocsql + "'"

	res := DBExec(sqlx)
	n, err := res.RowsAffected()
	checkerr(err)
	if n < 1 {
		fmt.Fprint(w, `{"res":"Deletion failed!"}`)
		return
	}

	fmt.Fprint(w, `{"res":"ok"}`)

}

func showlocations(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	if r.FormValue(Param_Labels["newloc"]) != "" {
		ajax_add_location(w, r)
		return
	}

	if r.FormValue(Param_Labels["delloc"]) != "" {
		ajax_del_location(w, r)
		return
	}

	start_html(w, r)

	sqlx := " FROM locations "

	NumLocations, _ := strconv.Atoi(getValueFromDB("SELECT Count(*) As rex "+sqlx, "0"))

	sqlx = " FROM locations LEFT JOIN (SELECT location AS xlocation,count(boxid) AS NumBoxes FROM boxes GROUP BY location) xlocations ON locations.location=xlocations.xlocation"

	sqllocation := ""
	if r.FormValue(Param_Labels["location"]) != "" {
		sqllocation, _ = url.QueryUnescape(r.FormValue(Param_Labels["location"]))
		sqllocation = strings.ReplaceAll(sqllocation, "'", "''")
		sqlx += " WHERE locations.location = '" + sqllocation + "'"

	}

	sqlx += " GROUP BY locations.location "
	if r.FormValue(Param_Labels["order"]) != "" && r.FormValue(Param_Labels["location"]) == "" {
		sqlx += "ORDER BY " + r.FormValue(Param_Labels["order"])
		if r.FormValue(Param_Labels["desc"]) != "" {
			sqlx += " DESC"
		}
	} else {
		sqlx += "ORDER BY locations.location"
	}

	flds := " id,locations.location, NumBoxes "
	if r.FormValue(Param_Labels["location"]) == "" {
		sqlx += emit_page_anchors(w, r, "locations", NumLocations)
	}
	//fmt.Printf("DEBUG: SELECT %v%v\n", flds, sqlx)
	printDebug("SELECT " + flds + " " + sqlx)
	rows, err := DBH.Query("SELECT  " + flds + sqlx)
	checkerr(err)
	defer rows.Close()

	var loc locationlistvars
	loc.Single = r.FormValue(Param_Labels["location"]) != ""
	loc.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"]) || r.FormValue(Param_Labels["order"]) == ""

	temp, err := template.New("locationListHead").Parse(templateLocationListHead)
	checkerr(err)
	err = temp.Execute(w, loc)
	checkerr(err)

	if runvars.Updating {
		temp, err = template.New("newLocation").Parse(templateNewLocation)
		checkerr(err)
		err = temp.Execute(w, "")
		checkerr(err)
	}
	temp, err = template.New("locationListLine2").Parse(templateLocationListLine)
	checkerr(err)
	for rows.Next() {
		loc.NumBoxes = 0
		rows.Scan(&loc.Id, &loc.Location, &loc.NumBoxes)
		loc.LocationUrl = url.QueryEscape(loc.Location)
		loc.NumBoxesX = commas(loc.NumBoxes)
		loc.DeleteOK = runvars.Updating && loc.NumBoxes == 0
		err := temp.Execute(w, loc)
		checkerr(err)
	}
	fmt.Fprint(w, ownerlisttrailer)

	if sqllocation != "" {
		showlocation(w, r, sqllocation, loc.NumBoxes)
	}

}

func showlocation(w http.ResponseWriter, r *http.Request, sqllocation string, NumBoxes int) {

	if r.FormValue(Param_Labels["location"]) == "" {
		show_search(w, r)
		return
	}

	var loc locationlistvars
	loc.Single = true
	loc.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])
	loc.Location, _ = url.QueryUnescape(r.FormValue(Param_Labels["location"]))
	loc.LocationUrl = r.FormValue(Param_Labels["location"])
	loc.NumBoxes, _ = strconv.Atoi(getValueFromDB("SELECT Count(*) As rex FROM boxes WHERE location='"+sqllocation+"'", "0"))
	loc.NumBoxesX = commas(loc.NumBoxes)
	loc.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"]) || r.FormValue(Param_Labels["order"]) == ""

	temp, err := template.New("locationBoxTableHead").Parse(templateLocationBoxTableHead)
	checkerr(err)
	err = temp.Execute(w, loc)
	checkerr(err)
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
	checkerr(err)
	defer rows.Close()
	var bv boxvars
	bv.Single = r.FormValue(Param_Labels["location"]) != ""
	bv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"]) || r.FormValue(Param_Labels["order"]) == ""

	temp, err = template.New("locationBoxTableRow").Parse(templateLocationBoxTableRow)
	checkerr(err)

	for rows.Next() {
		var mindate, maxdate string
		rows.Scan(&bv.Storeref, &bv.Boxid, &bv.Location, &bv.Contents, &bv.NumFiles, &mindate, &maxdate)
		if mindate == maxdate {
			bv.Date = mindate
			bv.ShowDate = formatShowDate(mindate)
			bv.Single = true
		} else {
			bv.Date = formatShowDate(mindate) + " to " + formatShowDate(maxdate)
			bv.ShowDate = bv.Date
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

	NumFiles, _ := strconv.Atoi(getValueFromDB("SELECT COUNT(*) AS rex FROM contents WHERE boxid='"+boxid+"'", "0"))
	sqllimit := emit_page_anchors(w, r, "boxes", NumFiles)
	sqlx := "SELECT owner,client,name,contents,review_date FROM contents "
	sqlx += " WHERE boxid='" + boxid + "'"

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

	temp, err := template.New("locationBoxFilesHead").Parse(templateLocationBoxFilesHead)
	checkerr(err)

	var bfv boxfilevars
	bfv.Boxid = boxid
	bfv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])

	err = temp.Execute(w, bfv)
	checkerr(err)

	nrows := 0
	temp, err = template.New("boxFilesLine").Parse(templateBoxFilesLine)
	checkerr(err)

	for rows.Next() {
		rows.Scan(&bfv.Owner, &bfv.Client, &bfv.Name, &bfv.Contents, &bfv.Date)
		bfv.OwnerUrl = template.URLQueryEscaper(bfv.Owner)
		bfv.ClientUrl = template.URLQueryEscaper(bfv.Client)
		err = temp.Execute(w, bfv)
		checkerr(err)

		nrows++
	}
	temp, err = template.New("boxfilestrailer").Parse(boxfilestrailer)
	temp.Execute(w, "")
	checkerr(err)

}

func default_location() string {

	sqlx := "SELECT location FROM locations ORDER BY location LIMIT 1"
	return getValueFromDB(sqlx, "")

}
