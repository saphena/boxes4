package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
)

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
	"userid":          "quu",
	"userpass":        "qup",
	"accesslevel":     "qal",
	"pagesize":        "qps",
	"offset":          "qof",
	"order":           "qor",
	"find":            "qqq",
	"desc":            "qds",
	"field":           "qfd",
	"overview":        "qov",
	"table":           "qtb",
}

type AppVars struct {
	Apptitle string
	Userid   string
}

var searchVars struct {
	Apptitle  string
	NumBoxes  int
	NumBoxesX string
	NumDocs   int
	NumDocsX  string
	NumLocns  int
	NumLocnsX string
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
}

const searchResultsHdr1 = `
<p>{{if .Find}}I was looking for <span class="searchedfor">{{if .Field}}{{.Field}} = {{end}}{{.Find}}</span> and{{end}} I found {{.Found}} matches.</p>
`

var searchResultsLine = `
<tr>
<td class="ourbox">{{if .Boxid}}<a href="/boxes?` + Param_Labels["boxid"] + `={{.BoxidUrl}}">{{end}}{{.Boxid}}{{if .Boxid}}</a>{{end}}</td>
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

const css = `
:root	{
	--regular-background	: #ffffe0;
	--regular-foreground	: black;
	--hilite-background		: yellow;
	--hilite-foreground		: black;
	--link-color			: blue;
	--link-hilite-fore		: black;
	--link-hilite-back		: orange;
}
body 				{
	background-color		: var(--regular-background);
	font-family				: Verdana, Arial, Helvetica, sans-serif; 
	font-size				: 16px; /*calc(8pt + 1vmin); */
	margin					: 1em;
	margin-top				: 6px;
	margin-bottom			: 6px;
}
a					{ 
						text-decoration: none;  
						color: var(--link-color); 
						font-family: monospace; 
						font-size: 1.2em;
						padding: 0 .5em 0 .5em; 
						/* &#8645; */
					}
a:visited			{ color: var(--link-color); }
a:hover             { 
						text-transform: uppercase; 
						font-weight: bold;
						color: var(--link-hilite-fore);
						background-color: var(--link-hilite-back);
					}
p.center			{ text-align: center; }
address 			{ font-size: 8pt; }
td 					{ padding: 4px; text-align: left; }
li					{ list-style-type: none; }

.pagelinks			{ padding: 2px 0 6px; 0 }
.numdocs,
.numboxes			{ text-align: center; }
.keydata			{ font-weight: bold; text-transform: uppercase; }

.btn {
	-webkit-border-radius: 5;
	-moz-border-radius: 5;
	border-radius: 5px;
	color: black;
	font-size: 20px;
	background: #9ea4a8;
	padding: 10px 20px 10px 20px;
	text-decoration: none;
  }
  
  .btn:hover {
	background: var(--link-hilite-back);
	color: var(--link-hilite-fore);
	text-decoration: none;
	text-transform: none;
  }
  
 
  td             		{ background: white; font-weight: bold; border-color: #bb0000; border-style: solid; border-width: 3px; }

td.center			{ text-align: center; }
td.left				{ text-align: left; }
td.right			{ text-align: right; }
td.category			{ background-color: #c8c8e8; font-weight: bold; }
td.col-1			{ background-color: #d8d8d8; }
td.col-2			{ background-color: #e8e8e8; }
td.form-title		{ background-color: #ffffff; font-weight: bold; }
td.nopad			{ padding: 0px; }
td.spacer			{ background-color: #ffffff; font-size: 1pt; line-height: 0.1; }
td.small-caption	{ font-size: 8pt; }
td.print			{ font-size: 8pt; text-align: center; }

tr.center			{ text-align: center; }
tr.row-1			{ background-color: #d8d8d8; }
tr.row-2			{ background-color: #e8e8e8; }
tr.spacer			{ background-color: #ffffff; }
tr.row-category		{ background-color: #c8c8e8; font-weight: bold; }

/* Login Info */
td.login-info-left	{ width: 33%; padding: 0px; text-align: left; }
td.login-info-middle{ width: 33%; padding: 0px; text-align: center; }
td.login-info-right	{ width: 33%; padding: 0px; text-align: right; }
span.login-username	{ font-style: italic; }
span.login-time		{ font-size: 8pt; font-style: italic; }

/* Menu */
td.menu				{ background-color: #e8e8e8; text-align: center; }

/* Quick Summary */
td.quick-summary-left	{ width: 50%; text-align: left; }
td.quick-summary-right	{ width: 50%; text-align: right; }

/* News */
td.news-heading		{ background-color: #d8d8d8; text-align: left; border-bottom: 1px solid #000000; }
td.news-body		{ background-color: #ffffff; padding: 16px; }
span.news-headline	{ font-weight: bold; }
span.news-date		{ font-style: italic; font-size: 8pt; }


th                  { text-align: left; padding: 2px;}
th.vertical         { text-align: right; font-weight: normal;}
h1                  { 
						text-align: center; 
						text-transform:uppercase; 
						/* color: blue; */
						text-shadow: -3px 3px 3px rgba(26, 205, 214, 1);
						}

.copyrite	{ font-size: small; }
.infohilite		{ background-color: yellow; color: black; font-weight: bold; padding-top: 4px; padding-bottom: 4px; }
.errormsg           { background-color: red; color: yellow; padding: 4px;}
.errordata          { background-color: red; color: white; font-weight: bold; font-size: larger; padding: 4px;}
.searchedfor		{ 
						background-color: var(--hilite-background); 
						color: var(--hilite-foreground); 
						font-weight: bold;
						font-size: larger;
						padding: 4px;
					}
.number             { text-align: right; }
.ourbox             { font-weight: bold; color: #bb0000; }
.upper				{ text-transform: uppercase; }
.lower				{ text-transform: lowercase; }

em	{font-style: italic; font-size: larger;}
.arrow				{ font-size: large; }
.topmenu 			{
	display: block;
	border-bottom: solid;
	padding-bottom: 3px;
	margin-bottom: 3px;
	width: 100%;
}
`

var html1 = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>{{.Apptitle}}{{if .Userid}}&#9997;{{end}}</title>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<script>
function isBadLength(sObj,iLen,sMsg) {

    if (sObj.value.length < iLen) {
      alert(sMsg)
      sObj.focus()
      return true
    }
  }

function changepagesize(sel) {
	let newpagesize = sel.value;
	let url = window.location.href;
	// Need to strip out any existing PAGESIZE
	let ps = url.match(/(&|\?)` + Param_Labels["pagesize"] + `\=\d+/);
	console.log('url="'+url+'"; ps="'+ps+'"; NP='+newpagesize);
	let cleanurl = url;
	if (ps) {
		cleanurl = cleanurl.replace(ps[0],'') + ps[1];
	} else {
		if (cleanurl.indexOf('?') < 0) {
			cleanurl += '?';
		} else {
			cleanurl += '&';
		}
	}
	console.log("cleanurl='"+cleanurl+"'");
	window.location.href = cleanurl + "` + Param_Labels["pagesize"] + `=" + newpagesize;
}
function trapkeys() {
	document.getElementsByTagName('body')[0].onkeyup = function(e) { 
		var ev = e || window.event;
	 	if (ev.keyCode == 37 || ev.keyCode == 33) { // Left arrow or PageUp
	   		let pp = document.getElementById('prevpage');
			if (pp) {
				window.location.href = pp.getAttribute('href');
			}
	   		return false;
		} else if (ev.keyCode == 39 || ev.keyCode == 34) { // Right arrow or PageDn
			let np = document.getElementById('nextpage');
		 	if (np) {
				window.location.href = np.getAttribute('href');
		 	}
			return false;
	    } 
	}

	let el = document.querySelector('body');
	swipedetect(el, function(swipedir){
		/* swipedir contains either "none", "left", "right", "top", or "down" */
		if (swipedir =='left') {
			console.log("swiped left");
			let pp = document.getElementById('prevpage');
			if (pp) {
				window.location.href = pp.getAttribute('href');
			}
		}
		else if (swipedir =='right') {
			alert("swiped right");
			let pp = document.getElementById('nextpage');
			if (pp) {
				window.location.href = pp.getAttribute('href');
			}
		}

	})
	



}


function swipedetect(el, callback){
  
    var touchsurface = el,
    swipedir,
    startX,
    startY,
    distX,
    distY,
    threshold = 150, //required min distance traveled to be considered swipe
    restraint = 100, // maximum distance allowed at the same time in perpendicular direction
    allowedTime = 300, // maximum time allowed to travel that distance
    elapsedTime,
    startTime,
    handleswipe = callback || function(swipedir){}
  
    touchsurface.addEventListener('touchstart', function(e){
        var touchobj = e.changedTouches[0]
        swipedir = 'none'
        dist = 0
        startX = touchobj.pageX
        startY = touchobj.pageY
        startTime = new Date().getTime() // record time when finger first makes contact with surface
        e.preventDefault()
    }, false)
  
    touchsurface.addEventListener('touchmove', function(e){
        e.preventDefault() // prevent scrolling when inside DIV
    }, false)
  
    touchsurface.addEventListener('touchend', function(e){
        var touchobj = e.changedTouches[0]
        distX = touchobj.pageX - startX // get horizontal dist traveled by finger while in contact with surface
        distY = touchobj.pageY - startY // get vertical dist traveled by finger while in contact with surface
        elapsedTime = new Date().getTime() - startTime // get time elapsed
        if (elapsedTime <= allowedTime){ // first condition for awipe met
            if (Math.abs(distX) >= threshold && Math.abs(distY) <= restraint){ // 2nd condition for horizontal swipe met
                swipedir = (distX < 0)? 'left' : 'right' // if dist traveled is negative, it indicates left swipe
            }
            else if (Math.abs(distY) >= threshold && Math.abs(distX) <= restraint){ // 2nd condition for vertical swipe met
                swipedir = (distY < 0)? 'up' : 'down' // if dist traveled is negative, it indicates up swipe
            }
        }
        handleswipe(swipedir)
        e.preventDefault()
    }, false)
}
  
//USAGE:
/*
var el = document.getElementById('someel')
swipedetect(el, function(swipedir){
    swipedir contains either "none", "left", "right", "top", or "down"
    if (swipedir =='left')
        alert('You just swiped left!')
})
*/



</script>

<style>

`

const html2 = `
-->
</style>
</head>
<body onload="trapkeys();">
<h1><a href="/">&#9783; {{.Apptitle}}</a> {{if .Userid}} <span style="font-size: 1.2em;" title="Update mode"> &#9997; </span>{{end}}</h1>
<div class="topmenu">
`

type locationlistvars struct {
	Location    string
	LocationUrl string
	Id          int
	NumBoxes    int
	NumBoxesX   string
	Desc        bool
	NumOrder    bool
	Single      bool
}

var locationlistline = `
<tr>
<td class="location">{{if .Single}}{{else}}{{if .Location}}<a href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}">{{end}}{{end}}{{.Location}}{{if .Single}}{{else}}{{if .Location}}</a>{{end}}{{end}}</td>
<td class="numboxes">{{.NumBoxesX}}</td>
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
<td class="number">{{.NumFilesX}}</td>
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
	Desc      bool
}

var ownerfilesline = `
<tr>
<td class="boxid">{{if .Boxid}}<a href="/boxes?` + Param_Labels["boxid"] + `={{.BoxidUrl}}">{{end}}{{.Boxid}}{{if .Boxid}}</a>{{end}}</td>
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
	//fmt.Printf("DEBUG: updating=%v usr=%v\n", updating, usr)
	if !updating {
		ht = html1 + css + html2 + basicMenu + "</div>"
	} else {
		if usr != nil {
			runvars.Userid = usr.(string)
		} else {
			runvars.Userid = ""
		}
		ht = html1 + css + html2 + updateMenu + "</div>"
	}
	html, err := template.New("mainmenu").Parse(ht)
	if err != nil {
		panic(err)
	}

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
	pagesizes := []int{0, 20, 40, 60, 100}
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
	AppTitle             string            `yaml:"AppTitle"`
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

FieldLabels:
  boxid:           'BoxID'
  owner:           'Owner'
  contents:        'Contents'
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
`
