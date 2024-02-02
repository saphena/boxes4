package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func exec_search(w http.ResponseWriter, r *http.Request) {

	// This needs to be here in order to collect runtime values from prefs

	start_html(w, r)

	var sqlx = ` FROM contents LEFT JOIN boxes ON contents.boxid=boxes.boxid `
	var wherex = ``
	var isDateSearch bool
	if r.FormValue(Param_Labels["find"]) != "" {
		if r.FormValue(Param_Labels["field"]) != "" {

			if r.FormValue(Param_Labels["field"]) == "review_date" {
				wherex += `review_date LIKE '?%'`
				isDateSearch = true
			} else if r.FormValue(Param_Labels["field"]) == "contents" {
				wherex += `contents LIKE '%?%'`
			} else if r.FormValue(Param_Labels["field"]) == "name" {
				wherex += `name LIKE '%?%'`
			} else if r.FormValue(Param_Labels["field"]) == "owner" {
				wherex += `(owner = '?' OR owner LIKE '?/%' OR owner LIKE '%/?')`
			} else if r.FormValue(Param_Labels["field"]) == "client" {
				wherex += `(client LIKE '?%' OR client LIKE '?/%' OR client LIKE '%/?')`
			} else {
				wherex += r.FormValue(Param_Labels["field"]) + `= '?'`
			}
		} else {
			wherex += `(
				(contents.boxid LIKE '?%')
			OR	(boxes.storeref LIKE '?%') 
			OR	(boxes.overview LIKE '%?%')
			OR	(contents.client LIKE '?%') 
			OR	(contents.client LIKE '?/%')
			OR	(contents.client LIKE '%/?')
			OR	(contents.owner = '?') 
			OR	(contents.owner LIKE '?/%')
			OR	(contents.owner LIKE '%/?')
			OR	(contents.contents LIKE '%?%') 
			OR	(contents.name LIKE '%?%')
			OR	(contents.review_date LIKE '?%')
			)
		`
		}
	}

	session, err := store.Get(r, cookie_name)
	checkerr(err)

	printDebug(fmt.Sprintf("%v", session.Values))

	if !isDateSearch {
		if prefs.IncludePastYears > 0 || session.Values["ExcludeBeforeYear"] != nil {
			firstyear := 0
			currentyear := time.Now().Year()
			if session.Values["ExcludeBeforeYear"] != nil {
				firstyear = session.Values["ExcludeBeforeYear"].(int)
			} else {
				firstyear = currentyear - prefs.IncludePastYears
			}
			if wherex != "" {
				wherex += " AND "
			}
			wherex += " review_date >= '" + strconv.Itoa(firstyear) + "'"
			printDebug(wherex)
		}
	}
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

	printDebug(sqlx)
	FoundRecCount, _ := strconv.Atoi(getValueFromDB("SELECT Count(*) AS Rexx"+sqlx, "0"))

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

	res.FindUrl = template.URLQueryEscaper(x)
	res.OneField = r.FormValue(Param_Labels["field"])

	res.Found = commas(FoundRecCount)
	res.Found0 = FoundRecCount == 0
	res.Found1 = FoundRecCount == 1
	res.Found2 = FoundRecCount > 1
	res.Field = prefs.Field_Labels[r.FormValue(Param_Labels["field"])]

	html, err := template.New("searchResultsHdr1").Parse(templateSearchResultsHdr1)
	checkerr(err)
	html.Execute(w, res)

	if res.Found0 {
		return
	}

	flds := "contents.boxid,contents.owner,contents.client,contents.name,contents.contents,contents.review_date,boxes.storeref,boxes.overview"

	sqllimit := emit_page_anchors(w, r, "find", FoundRecCount)

	rows, err := DBH.Query("SELECT " + flds + sqlx + sqllimit)
	if err != nil {
		fmt.Printf("Omg! %v\n", sqlx)
		panic(err)
	}
	html, err = template.New("searchResultsHdr2").Parse(templateSearchResultsHdr2)
	checkerr(err)
	html.Execute(w, res)

	html, err = template.New("searchResultsLine").Parse(templateSearchResultsLine)
	checkerr(err)
	for rows.Next() {
		rows.Scan(&res.Boxid, &res.Owner, &res.Client, &res.Name, &res.Contents, &res.Date, &res.Storeref, &res.Overview)
		res.BoxidUrl = template.URLQueryEscaper(res.Boxid)
		res.OwnerUrl = template.URLQueryEscaper(res.Owner)
		res.ClientUrl = template.URLQueryEscaper(res.Client)
		res.StorerefUrl = template.URLQueryEscaper(res.Storeref)
		res.ShowDate = formatShowDate(res.Date)
		res.DateYYMM = formatDateYYMM(res.Date)
		err = html.Execute(w, res)
		if err != nil {
			panic(err)
		}
	}
	fmt.Fprint(w, `</tbody></table>`)

	emitTrailer(w, r)

}

func show_search_params(w http.ResponseWriter, r *http.Request) {

	var params struct {
		Lrange    string
		Locations string

		Orange            string
		Owners            string
		Drange            string
		MaxYear           string
		ExcludeBeforeYear int
	}

	r.ParseForm()
	session, err := store.Get(r, cookie_name)
	checkerr(err)

	printDebug(fmt.Sprintf("%v\n", session.Values))
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

	if session.Values["ExcludeBeforeYear"] == nil {
		if prefs.IncludePastYears < 1 {
			params.Drange = Param_Labels["all"]
		} else {
			params.Drange = Param_Labels["selected"]
		}
		params.ExcludeBeforeYear = time.Now().Year() - prefs.IncludePastYears
		session.Values["ExcludeBeforeYear"] = params.ExcludeBeforeYear
	} else {
		eby := session.Values["ExcludeBeforeYear"].(int)
		if eby < 1 {
			params.Drange = Param_Labels["all"]
			params.ExcludeBeforeYear = time.Now().Year() - prefs.IncludePastYears
			session.Values["ExcludeBeforeYear"] = params.ExcludeBeforeYear
		} else {
			params.Drange = Param_Labels["selected"]
			params.ExcludeBeforeYear = session.Values["ExcludeBeforeYear"].(int)
		}
	}

	// Update settings

	if r.FormValue("l"+Param_Labels["range"]) != "" {
		params.Lrange = r.FormValue("l" + Param_Labels["range"])
		var locs []string
		for _, x := range r.Form[Param_Labels["location"]] {
			locs = append(locs, "'"+strings.ReplaceAll(x, "'", "''")+"'")
		}
		params.Locations = strings.Join(locs, ",")
		if params.Lrange == Param_Labels["all"] || len(params.Locations) == 0 {
			session.Values["locations"] = nil
		} else {
			session.Values["locations"] = params.Locations
		}
		err = store.Save(r, w, session)
		checkerr(err)
	}
	if r.FormValue("o"+Param_Labels["range"]) != "" {
		params.Orange = r.FormValue("o" + Param_Labels["range"])
		var owners []string
		for _, x := range r.Form[Param_Labels["owner"]] {
			owners = append(owners, "'"+strings.ReplaceAll(x, "'", "''")+"'")
		}
		params.Owners = strings.Join(owners, ",")
		if params.Orange == Param_Labels["all"] || len(params.Owners) == 0 {
			session.Values["owners"] = nil
		} else {
			session.Values["owners"] = params.Owners
		}
		err = store.Save(r, w, session)
		checkerr(err)
	}

	if r.FormValue("d"+Param_Labels["range"]) != "" {
		params.Drange = r.FormValue("d" + Param_Labels["range"])
		session.Values["ExcludeBeforeYear"] = 0
		if params.Drange == Param_Labels["selected"] {

			xby := r.FormValue(Param_Labels["ExcludeBeforeYear"])
			if xby != "" {
				xbyn, _ := strconv.Atoi(xby)
				if xbyn < 1 {
					session.Values["ExcludeBeforeYear"] = 0
				} else {
					session.Values["ExcludeBeforeYear"] = xbyn
				}
			}
		}
		err = store.Save(r, w, session)
		checkerr(err)
	}
	if r.FormValue(Param_Labels["savesettings"]) != "" {
		show_search(w, r)
		return
	}
	start_html(w, r)

	temp, err := template.New("searchParamsHead").Parse(templateSearchParamsHead)
	checkerr(err)
	temp.Execute(w, "")

	temp, err = template.New("searchParamsLocationRadios").Parse(templateSearchParamsLocationRadios)
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
	nmax := sessionPagesize(r)
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

	temp, err = template.New("searchParamsOwnerRadios").Parse(templateSearchParamsOwnerRadios)
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
	nmax = sessionPagesize(r)
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

	temp, err = template.New("searchParamsDateRadios").Parse(templateSearchParamsDateRadios)
	checkerr(err)

	params.MaxYear = getValueFromDB("SELECT ifnull(max(review_date),'2200') FROM contents", "2100")[0:4]
	params.ExcludeBeforeYear = 0
	if session.Values["ExcludeBeforeYear"] != nil {
		params.ExcludeBeforeYear = session.Values["ExcludeBeforeYear"].(int)
	} else if prefs.IncludePastYears > 0 {
		params.ExcludeBeforeYear = time.Now().Year() - prefs.IncludePastYears
	}

	err = temp.Execute(w, params)
	checkerr(err)
	fmt.Fprintln(w, "</div>")
	emitTrailer(w, r)
}

func show_search(w http.ResponseWriter, r *http.Request) {

	var sv searchVars

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
	sv.NumBoxes, _ = strconv.Atoi(getValueFromDB("SELECT Count(*) As Rex FROM boxes", "-1"))
	sv.NumBoxesX = commas(sv.NumBoxes)
	sv.NumDocs, _ = strconv.Atoi(getValueFromDB("SELECT Count(*) As Rex FROM contents", "-1"))
	sv.NumDocsX = commas(sv.NumDocs)
	sv.NumLocns, _ = strconv.Atoi(getValueFromDB("SELECT Count(*) As Rex FROM locations", "-1"))
	sv.NumLocnsX = commas(sv.NumLocns)
	sv.ExcludeBeforeYear = 0
	if session.Values["ExcludeBeforeYear"] != nil {
		sv.ExcludeBeforeYear = session.Values["ExcludeBeforeYear"].(int)
	} else if prefs.IncludePastYears > 0 {
		sv.ExcludeBeforeYear = time.Now().Year() - prefs.IncludePastYears
	}

	html, err := template.New("searchHome").Parse(templateSearchHome)
	checkerr(err)

	err = html.Execute(w, sv)
	checkerr(err)

	emitTrailer(w, r)
}
