package main

// This is the maximum number of pagelinks to show
// either side of the current page for paged lists
const MaxAdjacentPagelinks = 10

const ArrowPrevPage = `<span class="arrow">&#8666;</span>`
const ArrowNextPage = `<span class="arrow">&#8667;</span>`

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

const searchHTML = `
<p>Welcome to the {{.Apptitle}}.</p>

<p>I'm currently minding <strong>{{.NumDocs}}</strong> individual files packed into
<strong>{{.NumBoxes}}</strong> boxes stored in <strong>{{.NumLocns}}
</strong> locations.</p>

<form action="/find" onsubmit="return !isBadLength(this.FIND,1,'What would you have me look for illustrious one?');">
<input type="hidden" name="CMD" value="SEARCH"/>
<p>You can search the archives using a simple textsearch by entering the text you're looking for
here <input type="text" autofocus name="FIND"/><input type="submit" value="Find it!"/><br />
You can enter a partner's initials, a client number or name, a common term such as <em>tax</em> or a year.
Just enter the words you're looking for, no quote marks, ANDs, ORs, etc.</p></form>
<p>If you want to search only for records belonging to particular partners or locations, <a href="index.php?CMD=PARAMS">specify search options here</a>.</p>
<form action="/showbox"
    onsubmit="return !isBadLength(this.BOXID,1,
    'I\'m sorry, computers don\'t do guessing; you have to tell me which box to show you.\n\nPerhaps you want to see a list of boxes available in which case you should click on [boxes] above.');"><input type="hidden" name="CMD" value="SHOWBOX"/>
<p>If you want to look at a particular box, enter its ID here
<input type="text" name="BOXID" size="10"/><input type="submit" value="Show box"/></p></form>
`

type searchResultsVar struct {
	Boxid    string
	Partner  string
	Client   string
	Name     string
	Contents string
	Date     string
	Find     string
	Found    string
}

const searchResultsHdr = `
<p>{{if .Find}}I was looking for <span class="errordata">{{.Find}}</span> and{{end}} I found {{.Found}} matches.</p>
<table class="searchresults">
<thead>
<tr>
<th class="ourbox"><a href="/find?FIND={{.Find}}&ORDER=boxid{{if .Boxid}}&DESC=boxid{{end}}">Box ID</a></th>
<th class="owner"><a href="/find?FIND={{.Find}}&ORDER=owner{{if .Partner}}&DESC=owner{{end}}">Partner</a></th>
<th class="client"><a href="/find?FIND={{.Find}}&ORDER=client{{if .Client}}&DESC=client{{end}}">Client</a></th>
<th class="name"><a href="/find?FIND={{.Find}}&ORDER=name{{if .Name}}&DESC=name{{end}}">Name</a></th>
<th class="contents"><a href="/find?FIND={{.Find}}&ORDER=contents{{if .Contents}}&DESC=contents{{end}}">Contents</a></th>
<th class="date"><a href="/find?FIND={{.Find}}&ORDER=review_date{{if .Date}}&DESC=review_date{{end}}">Review</a></th>
</tr>
</thead>
<tbody>
`

const searchResultsLine = `
<tr>
<td class="ourbox"><a href="/showbox?BOXID={{.Boxid}}">{{.Boxid}}</a></td>
<td class="owner">{{.Partner}}</td>
<td class="client">{{.Client}}</td>
<td class="name">{{.Name}}</td>
<td class="contents">{{.Contents}}</td>
<td class="date">{{.Date}}</td>
</tr>
`

const searchResultsTrailer = `
</tbody>
</table>
`

const css = `
body 				{
	background-color: #FFFFE0;
	font-family:Verdana, Arial;
	font-size: 10pt;
	margin: 1em;
	margin-top: 6px;
	margin-bottom: 6px;
}
div.topmenu a       { text-decoration: none; }
div.pagelinks a		{ text-decoration: none; }
a:hover             { text-transform: uppercase; font-weight: bold; }
p 					{ font-family: Verdana, Arial, Helvetica; font-size: 10pt; }
p.center			{ text-align: center; }
address 			{ font-family: Verdana, Arial, Helvetica; font-size: 8pt; }
span				{ font-family: Verdana, Arial, Helvetica; font-size: 10pt; }

td 					{ font-family: Verdana, Arial, Helvetica;  padding: 4px; text-align: left; }
span.print			{ font-size: 8pt; }

span.required 		{ font-size: 8pt; color: #bb0000; }
span.small 			{ font-size: 8pt; }
span.pagetitle		{ font-size: 12pt; font-weight: bold; text-align: center; }
span.bold			{ font-weight: bold; }
span.italic			{ font-style: italic; }

table.hide			{ width: 100%; border-color: #ffffff; }
table.width100		{ width: 100%; border-color: #000000; border-style: solid; border-width: 1px; }
table.width75		{ width: 75%;  border-color: #000000; border-style: solid; border-width: 1px; }
table.width60		{ width: 60%;  border-color: #000000; border-style: solid; border-width: 1px; }
table.width50		{ width: 50%;  border-color: #000000; border-style: solid; border-width: 1px; }
table       		{              border-color: #bb0000; border-style: solid; border-width: 2px; }

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
h1                  { text-align: center; text-transform:uppercase; }

.copyrite	{ font-size: xx-small; }
.infohilite		{ background-color: yellow; color: black; font-weight: bold; padding-top: 4px; padding-bottom: 4px; }
.errormsg           { background-color: red; color: yellow; padding: 4px;}
.errordata          { background-color: red; color: white; font-weight: bold; font-size: larger; padding: 4px;}
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

const basicMenu = `
[<a href="/search" accesskey="s">search</a>] 
[<a href="/locations" accesskey="l">locations</a>] 
[<a href="/partners" accesskey="p">partners</a>] 
[<a href="/boxes" accesskey="b">boxes</a>] 
[<a href="/update" accesskey="u">update</a>] 
[<a href="/about" accesskey="a">about</a>] 

`

const updateMenu = `
[<a href="/search" accesskey="s">search</a>] 
[<a href="/locations" accesskey="l">locations</a>] 
[<a href="/partners" accesskey="p">partners</a>] 
[<a href="/boxes" accesskey="b">boxes</a>] 
[<a href="/users" accesskey="u">users</a>] 
[<a href="/logout" accesskey="l">logout {{.Username}</a>] 
[<a href="/about" accesskey="a">about</a>] 

`

const html1 = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>{{.Apptitle}}</title>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<script>
function changepagesize(sel) {
	let newpagesize = sel.value;
	let url = window.location.href;
	// Need to strip out any existing PAGESIZE
	let ps = url.match(/(&|\?)PAGESIZE=\d+/);
	console.log('url='+url+'; NP='+newpagesize);
	let cleanurl = url;
	if (ps) {
		cleanurl = url.replace(ps,'') + ps.substr(0,1);
	} else {
		if (cleanurl.indexOf('?') < 0) {
			cleanurl += '?';
		} else {
			cleanurl += '&';
		}
	}

	window.location.href = cleanurl + "PAGESIZE=" + newpagesize;
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
}
</script>

<style>

`
const html2 = `
-->
</style>
</head>
<body onload="trapkeys();">
<h1>{{.Apptitle}}</h1>
<div class="topmenu">
`

type partnerlistvars struct {
	Partner  string
	NumFiles int
	Desc     bool
	NumOrder bool
}

const partnerlisthdr = `
<table class="partnerlist">
<thead>
<tr>
<th class="partner"><a href="/partners?ORDER=owner{{if .Desc}}{{if .NumOrder}}{{else}}&DESC=owner{{end}}{{end}}">Partner</th>
<th class="number"><a href="/partners?ORDER=numdocs{{if .Desc}}{{if .NumOrder}}&DESC=numdocs{{end}}{{end}}">N<sup>o</sup> of files</th>
</tr>
</thead>
<tbody>
`

const partnerlistline = `
<tr>
<td class="partner">{{.Partner}}</td>
<td class="number">{{.NumFiles}}</td>
</tr>
`

const partnerlisttrailer = `
</tbody>
</table>
`

type boxvars struct {
	Boxid    string
	Location string
	Storeref string
	Contents string
	NumFiles int
	Date     string
}

const boxhtml = `
<table class="boxheader">
<tr><td class="vlabel">Box ID : </td><td class="vdata">{{.Boxid}}</td></tr>
<tr><td class="vlabel">Location : </td><td class="vdata"><a href="/showlocn?locn={{.Location}}">{{.Location}}</a></td></tr>
<tr><td class="vlabel">Storage ref : </td><td class="vdata">{{.Storeref}}</td></tr>
<tr><td class="vlabel">Contents : </td><td class="vdata">{{.Contents}}</td></tr>
<tr><td class="vlabel">N<sup>o</sup> of files : </td><td class="vdata">{{.NumFiles}}</td></tr>
<tr><td class="vlabel">Review date : </td><td class="vdata">{{.Date}}</td></tr>

</table>
`
const boxfileshdr = `
<table class="boxfiles">
<thead>
<tr>
<th class="owner">Partner</th>
<th class="client">Client</th>
<th class="name">Name</th>
<th class="contents">Contents</th>
<th class="date">Review</th>
</tr>
</thead>
<tbody>
`

type boxfilevars struct {
	Partner  string
	Client   string
	Name     string
	Contents string
	Date     string
}

const boxfilesline = `
<tr>
<td class="owner">{{.Partner}}</td>
<td class="client">{{.Client}}</td>
<td class="name">{{.Name}}</td>
<td class="contents">{{.Contents}}</td>
<td class="date">{{.Date}}</td>
</tr>
`
const boxfilestrailer = `
</tbody>
</table>
`
