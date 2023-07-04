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

	html, err := template.New("ownerListHead").Parse(templateOwnerListHead)
	checkerr(err)
	plv.Desc = r.FormValue(Param_Labels["desc"]) != r.FormValue(Param_Labels["order"])
	plv.NumOrder = r.FormValue(Param_Labels["order"]) == Param_Labels["numdocs"]
	err = html.Execute(w, plv)
	checkerr(err)

	html, err = template.New("ownerListLine").Parse(templateOwnerListLine)
	checkerr(err)
	for rows.Next() {
		rows.Scan(&plv.Owner, &plv.NumFiles)
		plv.OwnerUrl = template.URLQueryEscaper(plv.Owner)
		plv.NumFilesX = commas(plv.NumFiles)
		err := html.Execute(w, plv)
		checkerr(err)
	}
	fmt.Fprint(w, `</tbody></table>`)

	if owner == "" {
		emitTrailer(w, r)
		return
	}

	rows.Close()

	sqlx = " FROM contents  LEFT JOIN boxes ON contents.boxid=boxes.boxid "
	sqlx += " WHERE owner='" + strings.ReplaceAll(owner, "'", "''") + "'"
	NumRows, _ := strconv.Atoi(getValueFromDB("SELECT COUNT(*) AS rex"+sqlx, "0"))
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
	html, err = template.New("ownerFilesHead").Parse(templateOwnerFilesHead)
	checkerr(err)
	err = html.Execute(w, ofv)
	checkerr(err)

	html, err = template.New("ownerFilesLine").Parse(templateOwnerFilesLine)
	checkerr(err)
	for rows.Next() {
		rows.Scan(&ofv.Boxid, &ofv.Client, &ofv.Name, &ofv.Contents, &ofv.Date, &ofv.Overview)
		ofv.BoxidUrl = template.URLQueryEscaper(ofv.Boxid)
		ofv.ClientUrl = template.URLQueryEscaper(ofv.Client)
		ofv.ShowDate = formatShowDate((ofv.Date))
		err = html.Execute(w, ofv)
	}
	fmt.Fprint(w, `</tbody></table>`)
	emitTrailer(w, r)

}
