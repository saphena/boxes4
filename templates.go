package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// This is the maximum number of pagelinks to show
// either side of the current page for paged lists
const MaxAdjacentPagelinks = 10

const ArrowPrevPage = `<span class="arrow">&#8666;</span>`
const ArrowNextPage = `<span class="arrow">&#8667;</span>`

// Accesslevels are both discrete and hierarchical by numeric value
const ACCESSLEVEL_READONLY = 0
const ACCESSLEVEL_UPDATE = 2
const ACCESSLEVEL_SUPER = 9

var ACCESSLEVELS = map[int]string{
	ACCESSLEVEL_UPDATE:   "Can update",
	ACCESSLEVEL_SUPER:    "Controller",
	ACCESSLEVEL_READONLY: "View only",
}

var Field_Labels = map[string]string{
	"boxid":           "BoxID",
	"owner":           "Owner",
	"contents":        "Contents",
	"review_date":     "Review date",
	"name":            "Name",
	"client":          "Client",
	"location":        "Location",
	"numdocs":         "&#8470; of files",
	"min_review_date": "Min review date",
	"max_review_date": "Max review date",
	"userid":          "UserID",
	"userpass":        "Password",
	"accesslevel":     "Accesslevel",
	"storeref":        "Storage ref",
}

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
}

// Used to easily alter labels seen by user
var Menu_Labels = map[string]string{
	"search":    "search",
	"locations": "locations",
	"owners":    "owners",
	"boxes":     "boxes",
	"update":    "update",
	"logout":    "logout",
	"about":     "about",
}

type AppVars struct {
	Apptitle string
	Topmenu  string
}

var searchVars struct {
	Apptitle string
	NumBoxes int
	NumDocs  int
	NumLocns int
}

var searchHTML = `
<p>Welcome to the {{.Apptitle}}.</p>

<p>I'm currently minding <strong>{{.NumDocs}}</strong> individual files packed into
<strong>{{.NumBoxes}}</strong> boxes stored in <strong>{{.NumLocns}}
</strong> locations.</p>

<form action="/find">
<p>You can search the archives using a simple textsearch by entering the text you're looking for
here <input type="text" autofocus name="` + Param_Labels["find"] + `"/><input type="submit" value="Find it!"/><br />
You can enter a partner's initials, a client number or name, a common term such as <em>tax</em> or a year.
Just enter the words you're looking for, no quote marks, ANDs, ORs, etc.</p></form>
<p>If you want to search only for records belonging to particular partners or locations, <a href="index.php?CMD=PARAMS">specify search options here</a>.</p>
<form action="/boxes"
    onsubmit="return !isBadLength(this.` + Param_Labels["boxid"] + `,1,
    'I\'m sorry, computers don\'t do guessing; you have to tell me which box to show you.\n\nPerhaps you want to see a list of boxes available in which case you should click on [boxes] above.');">
<p>If you want to look at a particular box, enter its ID here
<input type="text" name="` + Param_Labels["boxid"] + `" size="10"/><input type="submit" value="Show box"/></p></form>
`

type searchResultsVar struct {
	Boxid    string
	Owner    string
	Client   string
	Name     string
	Contents string
	Date     string
	Find     string
	Found    string
	Desc     bool
	Storeref string
	Overview string
	Field    string
}

var searchResultsHdr1 = `
<p>{{if .Find}}I was looking for <span class="searchedfor">{{if .Field}}{{.Field}} = {{end}}{{.Find}}</span> and{{end}} I found {{.Found}} matches.</p>
`
var searchResultsHdr2 = `
<table class="searchresults">
<thead>
<tr>
<th class="ourbox"><a href="/find?` + Param_Labels["find"] + `={{.Find}}&` + Param_Labels["order"] + `=boxid{{if .Desc}}&` + Param_Labels["desc"] + `=boxid{{end}}">` + Field_Labels["boxid"] + `</a></th>
<th class="owner"><a href="/find?` + Param_Labels["find"] + `={{.Find}}&` + Param_Labels["order"] + `=owner{{if .Desc}}&` + Param_Labels["desc"] + `=owner{{end}}">` + Field_Labels["owner"] + `</a></th>
<th class="client"><a href="/find?` + Param_Labels["find"] + `={{.Find}}&` + Param_Labels["order"] + `=client{{if .Desc}}&` + Param_Labels["desc"] + `=client{{end}}">` + Field_Labels["client"] + `</a></th>
<th class="name"><a href="/find?` + Param_Labels["find"] + `={{.Find}}&` + Param_Labels["order"] + `=name{{if .Desc}}&` + Param_Labels["desc"] + `=name{{end}}">` + Field_Labels["name"] + `</a></th>
<th class="contents"><a href="/find?` + Param_Labels["find"] + `={{.Find}}&` + Param_Labels["order"] + `=contents{{if .Desc}}&` + Param_Labels["desc"] + `=contents{{end}}">` + Field_Labels["contents"] + `</a></th>
<th class="date"><a href="/find?` + Param_Labels["find"] + `={{.Find}}&` + Param_Labels["order"] + `=review_date{{if .Desc}}&` + Param_Labels["desc"] + `=review_date{{end}}">` + Field_Labels["review_date"] + `</a></th>
</tr>
</thead>
<tbody>
`

var searchResultsLine = `
<tr>
<td class="ourbox"><a href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}">{{.Boxid}}</a></td>
<td class="owner"><a href="/owners?` + Param_Labels["owner"] + `={{.Owner}}">{{.Owner}}</a></td>
<td class="client"><a href="/find?` + Param_Labels["find"] + `={{.Client}}&` + Param_Labels["field"] + `=client">{{.Client}}</a></td>
<td class="name">{{.Name}}</td>
<td class="contents">{{.Contents}}</td>
<td class="date"><a href="/find?` + Param_Labels["find"] + `={{.Date}}&` + Param_Labels["field"] + `=review_date{{.Date}}">{{.Date}}</a></td>
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
}
body 				{
	background-color		: var(--regular-background);
	font-family				: Verdana, Arial, Helvetica, sans-serif; 
	font-size				: 16px; /*calc(8pt + 1vmin); */
	margin					: 1em;
	margin-top				: 6px;
	margin-bottom			: 6px;
}
a					{ text-decoration: none; color: var(--link-color); }
a:visited			{ color: var(--link-color); }
a:hover             { text-transform: uppercase; font-weight: bold; }
p.center			{ text-align: center; }
address 			{ font-size: 8pt; }
td 					{ padding: 4px; text-align: left; }


.pagelinks			{ padding: 2px 0 6px; 0 }
.numdocs			{ text-align: center; }

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

.copyrite	{ font-size: xx-small; }
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

var basicMenu = `
[<a href="/search">` + Menu_Labels["search"] + `</a>] 
[<a href="/locations">` + Menu_Labels["locations"] + `</a>] 
[<a href="/owners">` + Menu_Labels["owners"] + `</a>] 
[<a href="/boxes">` + Menu_Labels["boxes"] + `</a>] 
[<a href="/update">` + Menu_Labels["update"] + `</a>] 
[<a href="/about">` + Menu_Labels["about"] + `</a>] 

`

var updateMenu = `
[<a href="/search">` + Menu_Labels["search"] + `</a>] 
[<a href="/locations">` + Menu_Labels["locations"] + `</a>] 
[<a href="/owners">` + Menu_Labels["owners"] + `</a>] 
[<a href="/boxes">` + Menu_Labels["boxes"] + `</a>] 
[<a href="/update">` + Menu_Labels["update"] + `</a>] 
[<a href="/about">` + Menu_Labels["about"] + `</a>] 
[<a href="/logout">` + Menu_Labels["logout"] + ` {{.Username}</a>] 

`

var html1 = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>{{.Apptitle}}</title>
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
<h1><a href="/">&#9783; {{.Apptitle}}</a></h1>
<div class="topmenu">
`

type ownerlistvars struct {
	Owner    string
	NumFiles int
	Desc     bool
	NumOrder bool
	Single   bool
}

var ownerlisthdr = `
<table class="ownerlist">
<thead>
<tr>


<th class="owner">{{if .Single}}{{else}}<a href="/owners?` + Param_Labels["order"] + `=owner{{if .Desc}}&` + Param_Labels["desc"] + `=owner{{end}}">{{end}}` + Field_Labels["owner"] + `{{if .Single}}{{else}}</a>{{end}}</th>
<th class="number">{{if .Single}}{{else}}<a href="/owners?` + Param_Labels["order"] + `=numdocs{{if .Desc}}&` + Param_Labels["desc"] + `=numdocs{{end}}">{{end}}` + Field_Labels["numdocs"] + `{{if .Single}}{{else}}</a>{{end}}</th>
</tr>
</thead>
<tbody>
`

var ownerlistline = `
<tr>
<td class="owner">{{if .Single}}{{else}}<a href="/owners?` + Param_Labels["owner"] + `={{.Owner}}">{{end}}{{.Owner}}{{if .Single}}{{else}}</a>{{end}}</td>
<td class="number">{{.NumFiles}}</td>
</tr>
`

const ownerlisttrailer = `
</tbody>
</table>
`

type ownerfilesvar struct {
	Owner    string
	Boxid    string
	Client   string
	Name     string
	Contents string
	Date     string
	Desc     bool
}

var ownerfileshdr = `
<table class="ownerfiles">
<thead>
<tr>

<th class="owner"><a href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=boxid{{if .Desc}}&` + Param_Labels["desc"] + `=boxid{{end}}">` + Field_Labels["boxid"] + `</a></th>
<th class="client"><a href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=client{{if .Desc}}&` + Param_Labels["desc"] + `=client{{end}}">` + Field_Labels["client"] + `</a></th>
<th class="name"><a href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=name{{if .Desc}}&` + Param_Labels["desc"] + `=name{{end}}">` + Field_Labels["name"] + `</a></th>
<th class="contents"><a href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=contents{{if .Desc}}&` + Param_Labels["desc"] + `=contents{{end}}">` + Field_Labels["contents"] + `</a></th>
<th class="review_date"><a href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=review_date{{if .Desc}}&` + Param_Labels["desc"] + `=review_date{{end}}">` + Field_Labels["review_date"] + `</a></th>
</tr>
</thead>
<tbody>
`

var ownerfilesline = `
<tr>
<td class="boxid"><a href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}">{{.Boxid}}</a></td>
<td class="client">{{if .Client}}<a href="/find?` + Param_Labels["find"] + `={{.Client}}&` + Param_Labels["field"] + `=client">{{end}}{{.Client}}{{if .Client}}</a>{{end}}</td>
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
	Location        string
	Storeref        string
	Contents        string
	NumFiles        int
	Overview        string
	Min_review_date string
	Max_review_date string
	Date            string
	Desc            bool
	Single          bool
}

var boxhtml = `
<table class="boxheader">


<tr><td class="vlabel">{{if .Single}}{{else}}<a href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=boxid&` + Param_Labels["desc"] + `=boxid">{{end}}` + Field_Labels["boxid"] + `{{if .Single}}{{else}}</a>{{end}} : </td><td class="vdata">{{.Boxid}}</td></tr>
<tr><td class="vlabel">` + Field_Labels["location"] + ` : </td><td class="vdata"><a href="/showlocn?` + Param_Labels["location"] + `={{.Location}}">{{.Location}}</a></td></tr>
<tr><td class="vlabel">` + Field_Labels["storeref"] + ` : </td><td class="vdata">{{.Storeref}}</td></tr>
<tr><td class="vlabel">` + Field_Labels["contents"] + ` : </td><td class="vdata">{{.Contents}}</td></tr>
<tr><td class="vlabel">` + Field_Labels["numdocs"] + ` : </td><td class="vdata">{{.NumFiles}}</td></tr>
<tr><td class="vlabel">` + Field_Labels["review_date"] + ` : </td><td class="vdata">{{.Date}}</td></tr>

</table>
`

var boxtablehdr = `
<table class="boxlist">
<thead>
<tr>


<th class="boxid"><a href="/boxes?` + Param_Labels["order"] + `=boxid{{if .Desc}}&` + Param_Labels["desc"] + `=boxid{{end}}">` + Field_Labels["boxid"] + `</a></th>
<th class="location"><a href="/boxes?` + Param_Labels["order"] + `=location{{if .Desc}}&` + Param_Labels["desc"] + `=location{{end}}">` + Field_Labels["location"] + `</a></th>
<th class="storeref"><a href="/boxes?` + Param_Labels["order"] + `=storeref{{if .Desc}}&` + Param_Labels["desc"] + `=storeref{{end}}">` + Field_Labels["storeref"] + `</a></th>
<th class="contents"><a href="/boxes?` + Param_Labels["order"] + `=contents{{if .Desc}}&` + Param_Labels["desc"] + `=contents{{end}}">` + Field_Labels["contents"] + `</a></th>
<th class="boxid"><a href="/boxes?` + Param_Labels["order"] + `=numdocs{{if .Desc}}&` + Param_Labels["desc"] + `=numdocs{{end}}">` + Field_Labels["numdocs"] + `</a></th>
<th class="boxid"><a href="/boxes?` + Param_Labels["order"] + `=min_review_date{{if .Desc}}&` + Param_Labels["desc"] + `=min_review_date{{end}}">` + Field_Labels["review_date"] + `</a></th>
</tr>
</thead>
<tbody>
`

var boxtablerow = `
<tr>
<td class="boxid"><a href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}">{{.Boxid}}</a></td>
<td class="location"><a href="/locations?` + Param_Labels["location"] + `={{.Location}}">{{.Location}}</a></td>
<td class="storeref"><a href="/find?` + Param_Labels["find"] + `={{.Storeref}}&` + Param_Labels["field"] + `=storeref">{{.Storeref}}</a></td>
<td class="overview">{{.Overview}}</td>
<td class="numdocs">{{.NumFiles}}</td>
<td class="review_date">{{if .Single}}<a href="find?` + Param_Labels["find"] + `={{.Date}}&` + Param_Labels["field"] + `=review_date">{{end}}{{.Date}}{{if .Single}}</a>{{end}}</td>
</tr>
`

var boxfileshdr = `
<table class="boxfiles">
<thead>
<tr>
<th class="owner"><a href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=owner{{if .Desc}}&` + Param_Labels["desc"] + `=owner{{end}}">` + Field_Labels["owner"] + `</a></th>
<th class="owner"><a href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=client{{if .Desc}}&` + Param_Labels["desc"] + `=client{{end}}">` + Field_Labels["client"] + `</a></th>
<th class="owner"><a href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=name{{if .Desc}}&` + Param_Labels["desc"] + `=name{{end}}">` + Field_Labels["name"] + `</a></th>
<th class="owner"><a href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=contents{{if .Desc}}&` + Param_Labels["desc"] + `=contents{{end}}">` + Field_Labels["contents"] + `</a></th>
<th class="owner"><a href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=review_date{{if .Desc}}&` + Param_Labels["desc"] + `=review_date{{end}}">` + Field_Labels["review_date"] + `</a></th>


</tr>
</thead>
<tbody>
`

type boxfilevars struct {
	Boxid    string
	Owner    string
	Client   string
	Name     string
	Contents string
	Date     string
	Desc     bool
}

var boxfilesline = `
<tr>
<td class="owner"><a href="/owners?` + Param_Labels["owner"] + `={{.Owner}}">{{.Owner}}</a></td>
<td class="client"><a href="/find?` + Param_Labels["find"] + `={{.Client}}&` + Param_Labels["field"] + `=client">{{.Client}}</a></td>
<td class="name">{{.Name}}</td>
<td class="contents">{{.Contents}}</td>
<td class="date"><a href="/find?` + Param_Labels["find"] + `={{.Date}}">{{.Date}}</a></td>
</tr>
`

const boxfilestrailer = `
</tbody>
</table>
`

func start_html(w http.ResponseWriter) {

	var ht string
	if true {
		ht = html1 + css + html2 + basicMenu + "</div>"
	} else {
		ht = html1 + css + html2 + updateMenu + "</div>"
	}
	html, err := template.New("main").Parse(ht)
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
		if k != "offset" && r.FormValue(v) != "" {
			if varx != "" {
				varx += "&"
			}
			varx += Param_Labels[k] + "=" + r.FormValue(v)
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
	if thisPage > MaxAdjacentPagelinks {
		minPage = thisPage - MaxAdjacentPagelinks
	}
	maxPage := numPages
	if thisPage < numPages-MaxAdjacentPagelinks {
		maxPage = thisPage + MaxAdjacentPagelinks
	}
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		if pageNum == 1 || pageNum == numPages || (pageNum >= minPage && pageNum <= maxPage) {
			if pageNum == thisPage {
				fmt.Fprintf(w, "[ <strong>%v</strong> ]", thisPage)
			} else {
				pOffset := (pageNum * pagesize) - pagesize

				fmt.Fprintf(w, `[<a href="/%v?%v`+Param_Labels["offset"]+`=%v" title="">%v</a>]`, cmd, varx, pOffset, strconv.Itoa(pageNum))
			}
		} else if pageNum == thisPage-(MaxAdjacentPagelinks+1) || pageNum == thisPage+MaxAdjacentPagelinks+1 {
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
