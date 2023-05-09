package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	_ "embed"
)

//go:embed boxes.js
var script string

const ArrowPrevPage = `<span class="arrow">&#8666;</span>`
const ArrowNextPage = `<span class="arrow">&#8667;</span>`

// Accesslevels are both discrete and hierarchical by numeric value
const ACCESSLEVEL_READONLY = 0
const ACCESSLEVEL_UPDATE = 2
const ACCESSLEVEL_SUPER = 9

// These labels, which must be unique, are used in URLs
// URL components might be found using RE matching, hence the 'q'
var Param_Labels = map[string]string{
	"boxid":           "qbx",
	"owner":           "qow",
	"contents":        "qcn",
	"review_date":     "qdt",
	"name":            "qnm",
	"client":          "qcl",
	"location":        "qlo",
	"numdocs":         "qnd",
	"min_review_date": "qd1",
	"max_review_date": "qd2",
	"userid":          "quu", // Hardcoded in boxes.js!
	"userpass":        "qup",
	"accesslevel":     "qal", // Hardcoded in boxes.js!
	"pagesize":        "qps", // Hardcoded in boxes.js!
	"offset":          "qof",
	"order":           "qor",
	"find":            "qqq",
	"desc":            "qds",
	"field":           "qfd",
	"overview":        "qov",
	"table":           "qtb",
	"textfile":        "qtx",
	"passchg":         "zpc", // Hardcoded in boxes.js!
	"single":          "z11", // Hardcoded in boxes.js!
	"multiple":        "z99",
	"oldpass":         "zop",
	"newpass":         "znp",
	"newpass2":        "z22",
	"adduser":         "zau",
	"deleteuser":      "zdu",
	"rowcount":        "zrc",
	"all":             "xal",
	"selected":        "xse",
	"range":           "xrg",
	"savesettings":    "sss",
	"newloc":          "nlc",
	"delloc":          "ndc",
	"newcontent":      "nct",
}

type AppVars struct {
	Apptitle string
	Userid   string
	Script   string
	Updating bool
}

type searchVars struct {
	Apptitle  string
	NumBoxes  int
	NumBoxesX string
	NumDocs   int
	NumDocsX  string
	NumLocns  int
	NumLocnsX string
	Locations string
	Owners    string
}

type searchResultsVar struct {
	Boxid       string
	BoxidUrl    string
	Owner       string
	OwnerUrl    string
	Client      string
	ClientUrl   string
	Name        string
	Contents    string
	Date        string
	Find        string
	FindUrl     string
	Found       string
	Desc        bool
	Storeref    string
	StorerefUrl string
	Overview    string
	Field       string
	Found0      bool
	Found1      bool
	Found2      bool
	Locations   string
	Owners      string
}

var searchResultsLine = `
<tr>
<td class="ourbox" title="{{.Overview}}">{{if .Boxid}}<a href="/boxes?` + Param_Labels["boxid"] + `={{.BoxidUrl}}">{{end}}{{.Boxid}}{{if .Boxid}}</a>{{end}}</td>
<td class="owner">{{if .Owner}}<a href="/owners?` + Param_Labels["owner"] + `={{.OwnerUrl}}">{{end}}{{.Owner}}{{if .Owner}}</a>{{end}}</td>
<td class="client">{{if .Client}}<a href="/find?` + Param_Labels["find"] + `={{.ClientUrl}}&` + Param_Labels["field"] + `=client">{{end}}{{.Client}}{{if .Client}}</a>{{end}}</td>
<td class="name">{{.Name}}</td>
<td class="contents">{{.Contents}}</td>
<td class="date">{{if .Date}}<a href="/find?` + Param_Labels["find"] + `={{.Date}}&` + Param_Labels["field"] + `=review_date">{{end}}{{.Date}}{{if .Date}}</a>{{end}}</td>
</tr>
`

const searchResultsTrailer = `
</tbody>
</table>
`

//go:embed boxes.css
var css string

var html1 = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>{{.Apptitle}}{{if .Userid}}&#9997;{{end}}</title>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<script>` + script + `</script>

<style>

`

const html2 = `
-->
</style>
</head>
<body onload="bodyLoaded();">
<h1><a href="/">&#9783; {{.Apptitle}}</a> {{if .Updating}} <span style="font-size: 1.2em;" title="Running in Update Mode"> &#9997; </span>{{end}}</h1>
<div class="topmenu">
`

const errormsgdiv = `<div id="errormsgdiv"></div>`

type locationlistvars struct {
	Location    string
	LocationUrl string
	Id          int
	NumBoxes    int
	NumBoxesX   string
	Desc        bool
	NumOrder    bool
	Single      bool
	DeleteOK    bool
}

var locationlistline = `
<tr>
<td class="location">{{if .Single}}{{else}}{{if .Location}}<a href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}">{{end}}{{end}}{{.Location}}{{if .Single}}{{else}}{{if .Location}}</a>{{end}}{{end}}</td>
<td class="numboxes">{{if .DeleteOK}}<input type="button" class="btn" value="Delete" onclick="delete_location(this);">{{else}}{{.NumBoxesX}}{{end}}</td>
</tr>
`

type ownerlistvars struct {
	Owner     string
	OwnerUrl  string
	NumFiles  int
	NumFilesX string
	Desc      bool
	NumOrder  bool
	Single    bool
}

var ownerlistline = `
<tr>
<td class="owner">{{if .Single}}{{else}}{{if .Owner}}<a href="/owners?` + Param_Labels["owner"] + `={{.OwnerUrl}}">{{end}}{{end}}{{.Owner}}{{if .Single}}{{else}}{{if .Owner}}</a>{{end}}{{end}}</td>
<td class="vdata">{{.NumFilesX}}</td>
</tr>
`

const ownerlisttrailer = `
</tbody>
</table>
`

type ownerfilesvar struct {
	Owner     string
	OwnerUrl  string
	Boxid     string
	BoxidUrl  string
	Client    string
	ClientUrl string
	Name      string
	Contents  string
	Date      string
	Overview  string
	Desc      bool
}

var ownerfilesline = `
<tr>
<td class="boxid" title="{{.Overview}}">{{if .Boxid}}<a href="/boxes?` + Param_Labels["boxid"] + `={{.BoxidUrl}}">{{end}}{{.Boxid}}{{if .Boxid}}</a>{{end}}</td>
<td class="client">{{if .Client}}<a href="/find?` + Param_Labels["find"] + `={{.ClientUrl}}&` + Param_Labels["field"] + `=client">{{end}}{{.Client}}{{if .Client}}</a>{{end}}</td>
<td class="name">{{.Name}}</td>
<td class="contents">{{.Contents}}</td>
<td class="review_date">{{if .Date}}<a href="/find?` + Param_Labels["find"] + `={{.Date}}&` + Param_Labels["field"] + `=review_date">{{end}}{{.Date}}{{if .Date}}</a>{{end}}</td>

</tr>
`

const ownerfilestrailer = `
</tbody>
</table>
`

type boxvars struct {
	Boxid           string
	BoxidUrl        string
	Location        string
	LocationUrl     string
	Storeref        string
	StorerefUrl     string
	Contents        string
	NumFiles        int
	NumFilesX       string
	Overview        string
	Min_review_date string
	Max_review_date string
	Date            string
	Desc            bool
	Single          bool
}

var boxtablerow = `
<tr>
<td class="boxid">{{if .Boxid}}<a href="/boxes?` + Param_Labels["boxid"] + `={{.BoxidUrl}}">{{end}}{{.Boxid}}{{if .Boxid}}</a>{{end}}</td>
<td class="location">{{if .Location}}<a href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}">{{end}}{{.Location}}{{if .Location}}</a>{{end}}</td>
<td class="storeref">{{if .Storeref}}<a href="/find?` + Param_Labels["find"] + `={{.StorerefUrl}}&` + Param_Labels["field"] + `=storeref">{{end}}{{.Storeref}}{{if .Storeref}}</a>{{end}}</td>
<td class="overview">{{.Overview}}</td>
<td class="numdocs">{{.NumFilesX}}</td>
<td class="review_date">{{if .Single}}{{if .Date}}<a href="find?` + Param_Labels["find"] + `={{.Date}}&` + Param_Labels["field"] + `=review_date">{{end}}{{end}}{{.Date}}{{if .Single}}{{if .Date}}</a>{{end}}{{end}}</td>
</tr>
`

// Header for box listing by location
var locboxtablerow = `
<tr>
<td class="boxid">{{if .Boxid}}<a href="/boxes?` + Param_Labels["boxid"] + `={{.BoxidUrl}}">{{end}}{{.Boxid}}{{if .Boxid}}</a>{{end}}</td>
<td class="storeref">{{if .Storeref}}<a href="/find?` + Param_Labels["find"] + `={{.StorerefUrl}}&` + Param_Labels["field"] + `=storeref">{{end}}{{.Storeref}}{{if .Storeref}}</a>{{end}}</td>
<td class="overview">{{.Contents}}</td>
<td class="numdocs">{{.NumFiles}}</td>
<td class="review_date">{{if .Single}}{{if .Date}}<a href="find?` + Param_Labels["find"] + `={{.Date}}&` + Param_Labels["field"] + `=review_date">{{end}}{{end}}{{.Date}}{{if .Single}}{{if .Date}}</a>{{end}}{{end}}</td>
</tr>
`

type boxfilevars struct {
	Boxid     string
	BoxidUrl  string
	Owner     string
	OwnerUrl  string
	Client    string
	ClientUrl string
	Name      string
	Contents  string
	Date      string
	Desc      bool
}

var boxfilesline = `
<tr>
<td class="owner">{{if .Owner}}<a href="/owners?` + Param_Labels["owner"] + `={{.OwnerUrl}}">{{end}}{{.Owner}}{{if .Owner}}</a>{{end}}</td>
<td class="client">{{if .Client}}<a href="/find?` + Param_Labels["find"] + `={{.ClientUrl}}&` + Param_Labels["field"] + `=client">{{end}}{{.Client}}{{if .Client}}</a>{{end}}</td>
<td class="name">{{.Name}}</td>
<td class="contents">{{.Contents}}</td>
<td class="date">{{if .Date}}<a href="/find?` + Param_Labels["find"] + `={{.Date}}">{{end}}{{.Date}}{{if .Date}}</a>{{end}}</td>
</tr>
`

func emit_owner_list(w http.ResponseWriter) {

	sqlx := "SELECT DISTINCT Trim(owner) FROM contents ORDER BY Trim(owner)"
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	fmt.Fprint(w, `<div class="hide"><datalist id="ownerlist">`)
	for rows.Next() {
		var owner string
		rows.Scan(&owner)
		fmt.Fprintf(w, `<option value="%v">`, owner)
	}
	fmt.Fprint(w, `</datalist></div>`)

}

func emit_client_list(w http.ResponseWriter) {

	sqlx := "SELECT DISTINCT Trim(client) FROM contents ORDER BY Trim(client)"
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	fmt.Fprint(w, `<div class="hide"><datalist id="clientlist">`)
	for rows.Next() {
		var client string
		rows.Scan(&client)
		fmt.Fprintf(w, `<option value="%v">`, client)
	}
	fmt.Fprint(w, `</datalist></div>`)

}

func emit_name_list(w http.ResponseWriter) {

	sqlx := "SELECT DISTINCT Trim(name) FROM contents ORDER BY Trim(name)"
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	fmt.Fprint(w, `<div class="hide"><datalist id="namelist">`)
	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Fprintf(w, `<option value="%v">`, name)
	}
	fmt.Fprint(w, `</datalist></div>`)

}

var newboxcontentline = `
<tr>
<td><input type="text" style="width:95%" list="ownerlist" class="keyinput"></td>
<td><input type="text" style="width:95%" list="clientlist" class="keyinput" oninput="fetch_client_name_list(this);"></td>
<td><input type="text" style="width:95%" list="namelist"></td>
<td><input type="text" style="width:95%"></td>
<td><input type="text" style="width:95%"></td>
<td><input type="button" class="btn" value="Add!" onclick="add_new_box_content(this);">
</tr>
`

const boxfilestrailer = `
</tbody>
</table>
`

func start_html(w http.ResponseWriter, r *http.Request) {
	var basicMenu = `
	[<a href="/search">` + prefs.Menu_Labels["search"] + `</a>] 
	[<a href="/locations">` + prefs.Menu_Labels["locations"] + `</a>] 
	[<a href="/owners">` + prefs.Menu_Labels["owners"] + `</a>] 
	[<a href="/boxes">` + prefs.Menu_Labels["boxes"] + `</a>] 
	[<a href="/update">` + prefs.Menu_Labels["update"] + `</a>] 
	[<a href="/about">` + prefs.Menu_Labels["about"] + `</a>] 
	
	`

	var updateMenu = `
	[<a href="/search">` + prefs.Menu_Labels["search"] + `</a>] 
	[<a href="/locations">` + prefs.Menu_Labels["locations"] + `</a>] 
	[<a href="/owners">` + prefs.Menu_Labels["owners"] + `</a>] 
	[<a href="/boxes">` + prefs.Menu_Labels["boxes"] + `</a>] 
	[<a href="/users">` + prefs.Menu_Labels["users"] + `</a>]
	[<a href="/logout">` + prefs.Menu_Labels["logout"] + ` {{.Userid}}</a>] 
	[<a href="/about">` + prefs.Menu_Labels["about"] + `</a>] 
	`

	var ht string

	updating, usr, _ := updateok(r)
	runvars.Updating = updating
	//fmt.Printf("DEBUG: updating=%v usr=%v\n", runvars.Updating, usr)
	if !runvars.Updating {
		ht = html1 + css + html2 + basicMenu + "</div>" + errormsgdiv
	} else {
		if usr != nil {
			runvars.Userid = usr.(string)
		} else {
			runvars.Userid = ""
		}
		ht = html1 + css + html2 + updateMenu + "</div>" + errormsgdiv
	}
	html, err := template.New("mainmenu").Parse(ht)
	checkerr(err)

	html.Execute(w, runvars)

}

func emit_page_anchors(w http.ResponseWriter, r *http.Request, cmd string, totrows int) string {

	pagesize := rangepagesize(r)
	offset := 0
	res := ""
	if pagesize > 0 {
		offset = rangeoffset(r)
		res = " LIMIT " + strconv.Itoa(offset)
		res += "," + strconv.Itoa(pagesize)
	}

	if pagesize < 1 {
		return res
	}
	numPages := totrows / pagesize
	if numPages*pagesize < totrows {
		numPages++
	}
	if numPages <= 1 {
		return res
	}
	varx := ""

	for k, v := range Param_Labels {
		varz, _ := url.QueryUnescape(r.FormValue(v))
		if k != "offset" && varz != "" {
			if varx != "" {
				varx += "&"
			}
			varx += Param_Labels[k] + "=" + url.QueryEscape(varz)
		}
	}
	if varx != "" {
		varx += "&"
	}

	fmt.Fprintf(w, `<div class="pagelinks">`)
	thisPage := (offset / pagesize) + 1
	if thisPage > 1 {
		prevPageOffset := (thisPage * pagesize) - (2 * pagesize)
		fmt.Fprintf(w, `&nbsp;&nbsp;<a id="prevpage" href="/%v?%v`+Param_Labels["offset"]+`=%v" title="Previous page">%v</a>&nbsp;&nbsp;`, cmd, varx, prevPageOffset, ArrowPrevPage)
	}
	minPage := 1
	if thisPage > prefs.MaxAdjacentPagelinks {
		minPage = thisPage - prefs.MaxAdjacentPagelinks
	}
	maxPage := numPages
	if thisPage < numPages-prefs.MaxAdjacentPagelinks {
		maxPage = thisPage + prefs.MaxAdjacentPagelinks
	}
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		if pageNum == 1 || pageNum == numPages || (pageNum >= minPage && pageNum <= maxPage) {
			if pageNum == thisPage {
				fmt.Fprintf(w, "[ <strong>%v</strong> ]", thisPage)
			} else {
				pOffset := (pageNum * pagesize) - pagesize

				fmt.Fprintf(w, `[<a href="/%v?%v`+Param_Labels["offset"]+`=%v" title="">%v</a>]`, cmd, varx, pOffset, strconv.Itoa(pageNum))
			}
		} else if pageNum == thisPage-(prefs.MaxAdjacentPagelinks+1) || pageNum == thisPage+prefs.MaxAdjacentPagelinks+1 {
			fmt.Fprintf(w, " ... ")
		}
	}
	if thisPage < numPages {
		nextPageOffset := (thisPage * pagesize)
		fmt.Fprintf(w, `&nbsp;&nbsp;<a id="nextpage" href="/%v?%v`+Param_Labels["offset"]+`=%v" title="Next page">%v</a>&nbsp;&nbsp;`, cmd, varx, nextPageOffset, ArrowNextPage)
	}

	fmt.Fprint(w, `<select onchange="changepagesize(this);">`)
	//pagesizes := []int{0, 20, 40, 60, 100}
	pagesizes := prefs.Pagesizes
	for _, ps := range pagesizes {
		fmt.Fprintf(w, `<option value="%v" `, ps)
		if ps == pagesize {
			fmt.Fprint(w, " selected ")
		}
		fmt.Fprint(w, `>`)
		if ps < 1 {
			fmt.Fprint(w, "show all")
		} else {
			fmt.Fprintf(w, "pagesize %v", ps)
		}
		fmt.Fprint(w, "</option>")
	}
	fmt.Fprint(w, "</select>")

	fmt.Fprintf(w, `</div>`)

	return res
}

type userpreferences struct {
	HttpPort             string            `yaml:"httpPort"`
	MaxAdjacentPagelinks int               `yaml:"MaxAdjacentPagelinks"`
	Accesslevels         map[int]string    `yaml:"AccesslevelNames"`
	MaxBoxContents       int               `yaml:"MaxBoxContents"`
	Field_Labels         map[string]string `yaml:"FieldLabels"`
	Menu_Labels          map[string]string `yaml:"MenuLabels"`
	Table_Labels         map[string]string `yaml:"TableLabels"`
	AppTitle             string            `yaml:"AppTitle"`
	CookieMaxAgeMins     int               `yaml:"LoginMinutes"`
	PasswordMinLength    int               `yaml:"PasswordMinLength"`
	DefaultPagesize      int               `yaml:"DefaultPagesize"`
	Pagesizes            []int             `yaml:"PagesizeOptions"`
	//pagesizes := []int{0, 20, 40, 60, 100}

}

const partial_config = `

httpPort: 4042

`

// YAML format configuration
const internal_config = `


AppTitle: 'document archives'

httpPort: 8081

# This is the maximum number of pagelinks to show
# either side of the current page for paged lists
MaxAdjacentPagelinks: 10

AccesslevelNames: 
  0: 'View only'
  2: 'Can update'
  9: 'Controller'


# Boxes containing more than this number of files are considered
# to be 'very large'
MaxBoxContents: 70

# This determines how long before a logged-in user is automatically logged out
# A value of 0 indicates logout whenever the browser session closes but this
# might not work as expected because of the browser's own settings
LoginMinutes: 60

PasswordMinLength: 4


# With no other info available, split things into chunks this big
DefaultPagesize: 20


PagesizeOptions: [0,20,40,60,100]

FieldLabels:
  boxid:           'BoxID'
  owner:           'Owner'
  contents:        'Files'
  review_date:     'Review date'
  name:            'Name'
  client:          'Client'
  location:        'Location'
  numdocs:         '&#8470; of files'
  numboxes:        '&#8470; of boxes'
  min_review_date: 'Min review date'
  max_review_date: 'Max review date'
  userid:          'UserID'
  userpass:        'Password'
  accesslevel:     'Accesslevel'
  storeref:        'Storage ref'
  overview:        'Contents'
  id:              'Id'

MenuLabels:
  search:    search
  locations: locations
  owners:    owners
  boxes:     boxes
  update:    update
  users:     users
  logout:    logout
  about:     about

TableLabels:
  boxes:		boxes
  contents:		files
  locations:	locations
  users:		users
  history:		history
`
