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

	if r.FormValue(Param_Labels["delbox"]) != "" {
		ajax_deleteEmptyBox(w, r)
		return
	}

	if r.FormValue(Param_Labels["newbox"]) != "" {
		ajax_create_new_box(w, r)
		return
	}

	if r.FormValue(Param_Labels["newok"]) != "" {
		ajax_check_new_boxid(w, r)
		return
	}

	if r.FormValue(Param_Labels["savebox"]) != "" {
		ajax_update_box_details(w, r)
		return
	}

	if r.FormValue(Param_Labels["chgboxlocn"]) != "" {
		ajax_changeBoxLocation(w, r)
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

	if r.FormValue(Param_Labels["boxid"]) != "" {
		showbox(w, r)
		return
	}

	start_html(w, r)

	sqlx := " FROM boxes "

	NumBoxes, _ := strconv.Atoi(getValueFromDB("SELECT Count(*) As rex "+sqlx, "0"))

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

	html, err := template.New("boxTableHead").Parse(templateBoxTableHead)
	checkerr(err)
	err = html.Execute(w, box)
	checkerr(err)

	if runvars.Updating {
		html, err = template.New("createNewBox").Parse(templateCreateNewBox)
		checkerr(err)
		err = html.Execute(w, "")
		checkerr(err)
	}
	html, err = template.New("boxTableRow").Parse(templateBoxTableRow)
	checkerr(err)
	for rows.Next() {
		rows.Scan(&box.Storeref, &box.Boxid, &box.Location, &box.Overview, &box.NumFiles, &box.Min_review_date, &box.Max_review_date)
		box.StorerefUrl = template.URLQueryEscaper(box.Storeref)
		box.BoxidUrl = template.URLQueryEscaper(box.Boxid)
		box.LocationUrl = template.URLQueryEscaper(box.Location)
		box.NumFilesX = commas(box.NumFiles)
		if box.Max_review_date == box.Min_review_date {
			box.Date = box.Max_review_date
			box.ShowDate = formatShowDate(box.Date)
			box.DateYYMM = formatDateYYMM(box.Date)
			box.Single = true
		} else {
			//fmt.Printf("Min date is %v, max date is %v\n", box.Min_review_date, box.Max_review_date)
			box.Date = formatShowDate(box.Min_review_date) + " to " + formatShowDate(box.Max_review_date)
			box.ShowDate = box.Date
			box.Single = false
		}
		err := html.Execute(w, box)
		checkerr(err)
	}
	fmt.Fprint(w, `</tbody></table>`)

	emitTrailer(w, r)

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
	checkerr(err)
	defer rows.Close()
	var bv boxvars
	bv.Single = r.FormValue(Param_Labels["boxid"]) != ""
	bv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])
	bv.UpdateOK = runvars.Updating
	bv.DeleteOK = runvars.Updating

	if !rows.Next() {
		fmt.Fprintf(w, "<p>No such box! %v</p>", r.FormValue(Param_Labels["boxid"]))
		return
	}
	//xx := prefs.Field_Labels["boxid"]
	//fmt.Printf("xx=%v\n", xx)
	var mindate, maxdate string
	rows.Scan(&bv.Storeref, &bv.Boxid, &bv.Location, &bv.Contents, &bv.NumFiles, &mindate, &maxdate)
	bv.StorerefUrl = template.URLQueryEscaper(bv.Storeref)
	if mindate == maxdate {
		bv.Date = formatShowDate(mindate)
	} else {
		bv.Date = formatShowDate(mindate) + " to " + formatShowDate(maxdate)
	}
	bv.NumFilesX = commas(bv.NumFiles)
	t := strings.ReplaceAll(templateBoxDetails, "#LOCSELECTOR#", generateLocationPicklist(bv.Location, "changeBoxLocation(this);"))
	html, err := template.New("boxDetails").Parse(t)
	checkerr(err)
	err = html.Execute(w, bv)
	checkerr(err)
	if runvars.Updating && bv.NumFiles < 1 {
		fmt.Fprint(w, `<div class="boxfunctionspanel">`)
		fmt.Fprintf(w, `<input type="button" value="Delete empty box" onclick="return deleteEmptyBox('%v');">`, r.FormValue(Param_Labels["boxid"]))
		fmt.Fprint(w, `</div>`)
	}
	showBoxfiles(w, r, sqlboxid)

}

func showBoxfiles(w http.ResponseWriter, r *http.Request, boxid string) {

	NumFiles, _ := strconv.Atoi(getValueFromDB("SELECT COUNT(*) AS rex FROM contents WHERE boxid='"+boxid+"'", "0"))
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

	html, err := template.New("boxFilesHead").Parse(templateBoxFilesHead)
	checkerr(err)

	var bfv boxfilevars
	bfv.Boxid = boxid
	bfv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])
	bfv.DeleteOK = runvars.Updating
	bfv.UpdateOK = runvars.Updating

	err = html.Execute(w, bfv)
	checkerr(err)

	if runvars.Updating {
		t := strings.ReplaceAll(templateNewBoxContentLine, "#DATESELECTORS#", generateDatePicklist(defaultReviewDate(), Param_Labels["review_date"], "newContentSaveNeeded(this.parentElement.parentElement);"))

		temp, err := template.New("newBoxContentLine").Parse(t)
		checkerr(err)
		err = temp.Execute(w, bfv)
		checkerr(err)
	}
	nrows := 0

	for rows.Next() {
		rows.Scan(&bfv.Owner, &bfv.Client, &bfv.Name, &bfv.Contents, &bfv.Date, &bfv.Id)
		bfv.OwnerUrl = template.URLQueryEscaper(bfv.Owner)
		bfv.ClientUrl = template.URLQueryEscaper(bfv.Client)

		t := strings.ReplaceAll(templateBoxFilesLine, "#DATESELECTORS#", generateDatePicklist(bfv.Date, Param_Labels["review_date"], "contentSaveNeeded(this.parentElement);"))
		bfv.ShowDate = formatShowDate(bfv.Date)
		bfv.DateYYMM = formatDateYYMM(bfv.Date)
		html, err = template.New("boxFilesLine").Parse(t)
		checkerr(err)

		err = html.Execute(w, bfv)
		checkerr(err)

		nrows++
	}
	fmt.Fprint(w, `</tbody></table>`)

	emit_owner_list(w)
	emit_client_list(w)
	emit_name_list(w)
	emitTrailer(w, r)

}

func ajax_fetch_name_list(w http.ResponseWriter, r *http.Request) {

	client := r.FormValue(Param_Labels["client"])
	sqlx := "SELECT DISTINCT Trim(name) FROM contents"
	if client != "" {
		sqlx += " WHERE client='" + strings.ReplaceAll(client, "'", "''") + "'"
	}
	sqlx += " ORDER BY Trim(name)"
	printDebug(sqlx)
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
	review := r.FormValue(Param_Labels["review_date"])

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

	printDebug(sqlx)
	res := DBExec(sqlx)
	printDebug(fmt.Sprintf("Result is %v\n", res))
	if res == nil {
		printDebug(`{"res":"Database operation failed!"}`)
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}
	n, _ := res.RowsAffected()
	//checkerr(err)
	if n < 1 {
		printDebug(`{"res":"Database operation failed!"}`)
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}
	n, err := res.LastInsertId()
	checkerr(err)
	syncOwner(owner)
	nf, ld, hd := update_ajax_box_contents(boxid)
	fmt.Fprintf(w, `{"res":"ok","nfiles":"%v","lodate":"%v","hidate":"%v","recid":"%v"}`, nf, ld, hd, n)

}

func ajax_delete_content_line(w http.ResponseWriter, r *http.Request) {

	id := r.FormValue(Param_Labels["delcontent"])
	owner := r.FormValue(Param_Labels["owner"])
	client := r.FormValue(Param_Labels["client"])
	boxid := r.FormValue(Param_Labels["boxid"])

	sqlx := "DELETE FROM contents WHERE id=" + id
	sqlx += " AND owner='" + safesql(owner) + "'"
	sqlx += " AND client='" + safesql(client) + "'"
	printDebug(sqlx)
	res := DBExec(sqlx)
	if res == nil {
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}
	n, err := res.RowsAffected()
	checkerr(err)
	if n < 1 {
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
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

	printDebug(sqlx)
	res := DBExec(sqlx)
	if res == nil {
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}
	n, err := res.RowsAffected()
	checkerr(err)
	if n < 1 {
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}

	syncOwner(owner)

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
	if nfiles < 1 {
		hidate = lodate
	}
	sqlx = "UPDATE boxes SET numdocs=" + strconv.Itoa(nfiles)
	sqlx += ",min_review_date='" + lodate + "'"
	sqlx += ",max_review_date='" + hidate + "'"
	sqlx += "WHERE boxid='" + safesql(boxid) + "'"
	printDebug(sqlx)
	DBExec(sqlx)

	return nfiles, lodate, hidate

}

func ajax_check_new_boxid(w http.ResponseWriter, r *http.Request) {

	boxid := r.FormValue(Param_Labels["newok"])

	sqlx := "SELECT boxid FROM boxes WHERE boxid='" + safesql(boxid) + "'"
	if len(boxid) < 1 || getValueFromDB(sqlx, "") != "" {
		printDebug("Replying boxid exists")
		fmt.Fprint(w, `{"res":"Duplicate!"}`)
	} else {
		printDebug("Replying boxid ok")
		fmt.Fprint(w, `{"res":"ok"}`)
	}
}

func ajax_update_box_details(w http.ResponseWriter, r *http.Request) {

	boxid := r.FormValue(Param_Labels["savebox"])
	storeref := r.FormValue(Param_Labels["storeref"])
	overview := r.FormValue(Param_Labels["overview"])

	sqlx := "UPDATE boxes SET "
	sqlx += " storeref='" + safesql(storeref) + "'"
	sqlx += ", overview='" + safesql(overview) + "'"
	sqlx += " WHERE boxid='" + safesql(boxid) + "'"

	res := DBExec(sqlx)
	if res == nil {
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}
	n, err := res.RowsAffected()
	checkerr(err)
	if n < 1 {
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}

	fmt.Fprint(w, `{"res":"ok"}`)

}

func ajax_changeBoxLocation(w http.ResponseWriter, r *http.Request) {

	boxid := r.FormValue(Param_Labels["boxid"])
	locn := r.FormValue(Param_Labels["chgboxlocn"])

	sqlx := "SELECT location FROM locations WHERE location='" + safesql(locn) + "'"
	if getValueFromDB(sqlx, "") == "" {
		fmt.Fprint(w, `{"res":"`+prefs.Field_Labels["location"]+` doesn't exist"}`)
		return
	}
	sqlx = "UPDATE boxes SET location='" + safesql(locn) + "' WHERE boxid='" + safesql(boxid) + "'"
	printDebug(sqlx)
	res := DBExec(sqlx)
	if res == nil {
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}
	n, err := res.RowsAffected()
	checkerr(err)
	if n < 1 {
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}

	fmt.Fprint(w, `{"res":"ok"}`)
}

func ajax_deleteEmptyBox(w http.ResponseWriter, r *http.Request) {

	boxid := r.FormValue(Param_Labels["delbox"])

	sqlx := "DELETE FROM boxes WHERE boxid='" + safesql(boxid) + "'"

	printDebug(sqlx)
	res := DBExec(sqlx)
	if res == nil {
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}
	n, err := res.RowsAffected()
	checkerr(err)
	if n < 1 {
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}

	fmt.Fprint(w, `{"res":"ok"}`)

}
func ajax_create_new_box(w http.ResponseWriter, r *http.Request) {

	boxid := r.FormValue(Param_Labels["newbox"])

	sqlx := "SELECT boxid FROM boxes WHERE boxid='" + safesql(boxid) + "'"
	if getValueFromDB(sqlx, "") != "" {
		fmt.Fprint(w, `{"res":"`+prefs.Field_Labels["boxid"]+` already exists"}`)
		return
	}
	sqlx = "INSERT INTO boxes (boxid,location,storeref,overview,numdocs) VALUES("
	sqlx += "'" + safesql(boxid) + "'"
	sqlx += ",'" + safesql(default_location()) + "'"
	sqlx += ",'" + safesql(boxid) + "'"
	sqlx += ",'" + prefs.Literals["newboxoverview"] + "'"
	sqlx += ",0"
	sqlx += ")"

	printDebug(sqlx)
	res := DBExec(sqlx)
	if res == nil {
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}
	n, err := res.RowsAffected()
	checkerr(err)
	if n < 1 {
		fmt.Fprint(w, `{"res":"Database operation failed!"}`)
		return
	}

	fmt.Fprint(w, `{"res":"ok"}`)
}
