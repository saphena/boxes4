package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	_ "embed"
)

//go:embed boxes.js
var script string

const ArrowPrevPage = `<span class="arrow">&#8666;</span>`
const ArrowNextPage = `<span class="arrow">&#8667;</span>`

// If a date is invalid, replace it with this value
const InvalidDateValue = "2000-01-01"

// Accesslevels are both discrete and hierarchical by numeric value
const ACCESSLEVEL_READONLY = 0
const ACCESSLEVEL_UPDATE = 2
const ACCESSLEVEL_SUPER = 9

// These labels, which must be unique, are used in URLs
// URL components might be found using RE matching, hence the 'q'
var Param_Labels = map[string]string{
	"boxid":             "qbx",
	"owner":             "qow",
	"contents":          "qcn",
	"review_date":       "qdt",
	"name":              "qnm",
	"client":            "qcl",
	"location":          "qlo",
	"storeref":          "qlr",
	"numdocs":           "qnd",
	"min_review_date":   "qd1",
	"max_review_date":   "qd2",
	"userid":            "quu",
	"userpass":          "qup",
	"accesslevel":       "qal",
	"pagesize":          "qps",
	"offset":            "qof",
	"order":             "qor",
	"find":              "qqq",
	"desc":              "qds",
	"field":             "qfd",
	"overview":          "qov",
	"table":             "qtb",
	"textfile":          "qtx",
	"passchg":           "zpc",
	"single":            "z11",
	"multiple":          "z99",
	"oldpass":           "zop",
	"newpass":           "znp",
	"newpass2":          "z22",
	"adduser":           "zau",
	"deleteuser":        "zdu",
	"rowcount":          "zrc",
	"all":               "xal",
	"selected":          "xse",
	"range":             "xrg",
	"savesettings":      "sss",
	"newloc":            "nlc",
	"delloc":            "ndc",
	"newcontent":        "nct",
	"delcontent":        "dct",
	"savecontent":       "dsc",
	"chgboxlocn":        "dxl",
	"savebox":           "dbx",
	"newok":             "xid",
	"newbox":            "xnb",
	"delbox":            "kbx",
	"ExcludeBeforeYear": "xby",
	"theme":             "ttt",
	"delowner":          "kwn",
}

type AppVars struct {
	Apptitle string
	Userid   string
	Script   string
	Updating bool
}

type searchVars struct {
	Apptitle          string
	NumBoxes          int
	NumBoxesX         string
	NumDocs           int
	NumDocsX          string
	NumLocns          int
	NumLocnsX         string
	Locations         string
	Owners            string
	ExcludeBeforeYear int
}

type searchResultsVar struct {
	Boxid             string
	BoxidUrl          string
	Owner             string
	OwnerUrl          string
	Client            string
	ClientUrl         string
	Name              string
	Contents          string
	Date              string
	ShowDate          string
	DateYYMM          string
	Find              string
	FindUrl           string
	Found             string
	OneField          string
	Desc              bool
	Storeref          string
	StorerefUrl       string
	Overview          string
	Field             string
	Found0            bool
	Found1            bool
	Found2            bool
	Locations         string
	Owners            string
	ExcludeBeforeYear int
}

var templateSearchHome string
var templateSearchResultsHdr1 string
var templateSearchResultsHdr2 string
var templateSearchResultsLine string
var templateSearchParamsHead string
var templateSearchParamsLocationRadios string
var templateSearchParamsOwnerRadios string
var templateSearchParamsDateRadios string
var templateLocationBoxTableHead string
var templateLocationListHead string
var templateLocationListLine string
var templateNewLocation string

var templateOwnerListHead string
var templateOwnerListLine string
var templateOwnerFilesHead string
var templateOwnerFilesLine string
var templateBoxTableHead string
var templateBoxTableRow string
var templateLocationBoxTableRow string
var templateCreateNewBox string
var templateBoxDetails string
var templateBoxFilesHead string
var templateBoxFilesLine string
var templateNewBoxContentLine string
var templateUserLoginHome string
var templateUserPasswordChange string
var templateMultiUserPasswordChangeHead string
var templateMultiUserPasswordChangeLine string
var templateMultiUserPasswordChangeFoot string

func initTemplates() {

	initSearchTemplates()
	initLocationTemplates()
	initOwnerTemplates()
	initBoxTemplates()
	initUserTemplates()

} // initTemplates

func initBoxTemplates() {

	templateBoxTableHead = `
<table class="boxlist">
<thead>
<tr>


<th class="boxid"><a title="&#8645;" class="sortlink" href="/boxes?` + Param_Labels["order"] + `=boxid{{if .Desc}}&` + Param_Labels["desc"] + `=boxid{{end}}">` + prefs.Field_Labels["boxid"] + `</a></th>
<th class="location"><a title="&#8645;" class="sortlink" href="/boxes?` + Param_Labels["order"] + `=location{{if .Desc}}&` + Param_Labels["desc"] + `=location{{end}}">` + prefs.Field_Labels["location"] + `</a></th>
<th class="storeref"><a title="&#8645;" class="sortlink" href="/boxes?` + Param_Labels["order"] + `=storeref{{if .Desc}}&` + Param_Labels["desc"] + `=storeref{{end}}">` + prefs.Field_Labels["storeref"] + `</a></th>
<th class="contents"><a title="&#8645;" class="sortlink" href="/boxes?` + Param_Labels["order"] + `=overview{{if .Desc}}&` + Param_Labels["desc"] + `=overview{{end}}">` + prefs.Field_Labels["overview"] + `</a></th>
<th class="boxid"><a title="&#8645;" class="sortlink" href="/boxes?` + Param_Labels["order"] + `=numdocs{{if .Desc}}&` + Param_Labels["desc"] + `=numdocs{{end}}">` + prefs.Field_Labels["numdocs"] + `</a></th>
<th class="date center"><a title="&#8645;" class="sortlink" href="/boxes?` + Param_Labels["order"] + `=min_review_date{{if .Desc}}&` + Param_Labels["desc"] + `=min_review_date{{end}}">` + prefs.Field_Labels["review_date"] + `</a></th>
</tr>
</thead>
<tbody>
`

	templateBoxDetails = `
<input type="hidden" id="AutosaveSeconds" value="` + strconv.Itoa(prefs.AutosaveSeconds) + `">
<table class="boxheader">


<tr><td class="vlabel">{{if .Single}}{{else}}<a title="&#8645;" href="/boxes?` + Param_Labels["boxid"] + `={{.BoxidUrl}}&` + Param_Labels["order"] + `=boxid&` + Param_Labels["desc"] + `=boxid">{{end}}` + prefs.Field_Labels["boxid"] + `{{if .Single}}{{else}}</a>{{end}} : </td><td id="boxboxid" class="vdata">{{.Boxid}}</td>
{{if .UpdateOK}}
<td class="nude"><input type="button" class="btn hide" id="updateboxbutton" value="Save changes" data-boxid="{{.Boxid}}" onclick="updateBoxDetails(this);">
{{end}}
</tr>

<tr>
<td class="vlabel">` + prefs.Field_Labels["location"] + ` : </td>
<td class="vdata">{{if .UpdateOK}}#LOCSELECTOR#{{else}}<a href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}">{{.Location}}</a>{{end}}</td>
</tr>

<tr><td class="vlabel">` + prefs.Field_Labels["storeref"] + ` : </td>
<td class="vdata" id="boxstoreref"{{if .UpdateOK}} contenteditable="true" oninput="boxDetailsSaveNeeded(this);">{{.Storeref}}{{else}}><a class="lookuplink" title="Search for {{.Storeref}}" href="/find?` + Param_Labels["find"] + `={{.StorerefUrl}}&` + Param_Labels["field"] + `=storeref">{{.Storeref}}</a>{{end}}</td></tr>

<tr><td class="vlabel">` + prefs.Field_Labels["overview"] + ` : </td>
<td class="vdata" id="boxoverview"{{if .UpdateOK}} contenteditable="true" oninput="boxDetailsSaveNeeded(this);"{{end}}>{{.Contents}}</td></tr>

<tr><td class="vlabel">` + prefs.Field_Labels["numdocs"] + ` : </td><td id="boxnumfiles" class="vdata numdocs">{{.NumFilesX}}</td></tr>
<tr><td class="vlabel">` + prefs.Field_Labels["review_date"] + ` : </td><td id="boxdates" class="vdata center">{{.Date}}</td></tr>

</table>
`

	templateBoxFilesHead = `
<table class="boxfiles">
<thead>
<tr>
<th class="owner"><a title="&#8645;" class="sortlink" href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=owner{{if .Desc}}&` + Param_Labels["desc"] + `=owner{{end}}">` + prefs.Field_Labels["owner"] + `</a></th>
<th class="owner"><a title="&#8645;" class="sortlink" href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=client{{if .Desc}}&` + Param_Labels["desc"] + `=client{{end}}">` + prefs.Field_Labels["client"] + `</a></th>
<th class="owner"><a title="&#8645;" class="sortlink" href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=name{{if .Desc}}&` + Param_Labels["desc"] + `=name{{end}}">` + prefs.Field_Labels["name"] + `</a></th>
<th class="owner"><a title="&#8645;" class="sortlink" href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=contents{{if .Desc}}&` + Param_Labels["desc"] + `=contents{{end}}">` + prefs.Field_Labels["contents"] + `</a></th>
<th class="owner"><a class="sortlink" href="/boxes?` + Param_Labels["boxid"] + `={{.Boxid}}&` + Param_Labels["order"] + `=review_date{{if .Desc}}&` + Param_Labels["desc"] + `=review_date{{end}}">` + prefs.Field_Labels["review_date"] + `</a></th>


</tr>
</thead>
<tbody>
`

	templateBoxFilesLine = `
<tr data-id="{{.Id}}">
<td class="owner" {{if .UpdateOK}}contenteditable="true" oninput="contentSaveNeeded(this);">{{.Owner}}{{else}}>{{if .Owner}}<a class="lookuplink" href="/owners?` + Param_Labels["owner"] + `={{.OwnerUrl}}">{{end}}{{.Owner}}{{if .Owner}}</a>{{end}}{{end}}</td>
<td class="client" {{if .UpdateOK}}contenteditable="true" oninput="contentSaveNeeded(this);">{{.Client}}{{else}}>{{if .Client}}<a class="lookuplink" href="/find?` + Param_Labels["find"] + `={{.ClientUrl}}&` + Param_Labels["field"] + `=client">{{end}}{{.Client}}{{if .Client}}</a>{{end}}{{end}}</td>
<td class="name" {{if .UpdateOK}}contenteditable="true" oninput="contentSaveNeeded(this);"{{end}}>{{.Name}}</td>
<td class="contents" {{if .UpdateOK}}contenteditable="true" oninput="contentSaveNeeded(this);"{{end}}>{{.Contents}}</td>
{{if .UpdateOK}}
<td class="date center">
#DATESELECTORS#
</td>
{{else}}
<td class="date center">{{if .Date}}<a class="lookuplink" href="/find?` + Param_Labels["find"] + `={{.DateYYMM}}&` + Param_Labels["field"] + `=review_date">{{end}}{{.ShowDate}}{{if .Date}}</a>
{{end}}
{{end}}</td>
{{if .UpdateOK}}<td class="center">
<input type="button" class="btn hide" data-id="{{.Id}}" data-boxid="{{.Boxid}}" value="Save changes" onclick="updateBoxContentLine(this);">
{{if .DeleteOK}}
<input type="checkbox" title="Enable delete button" onchange="this.nextElementSibling.classList.remove('hide');this.classList.add('hide');">
<input type="button" class="btn hide" data-id="{{.Id}}" data-boxid="{{.Boxid}}" value="Delete" onclick="deleteBoxContentLine(this);">
{{end}}
</td>{{end}}
</tr>
`

	templateBoxTableRow = `
<tr>
<td class="boxid">{{if .Boxid}}<a href="/boxes?` + Param_Labels["boxid"] + `={{.BoxidUrl}}">{{end}}{{.Boxid}}{{if .Boxid}}</a>{{end}}</td>
<td class="location">{{if .Location}}<a href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}">{{end}}{{.Location}}{{if .Location}}</a>{{end}}</td>
<td class="storeref">{{if .Storeref}}<a class="lookuplink" title="Search for {{.Storeref}}" href="/find?` + Param_Labels["find"] + `={{.StorerefUrl}}&` + Param_Labels["field"] + `=storeref">{{end}}{{.Storeref}}{{if .Storeref}}</a>{{end}}</td>
<td class="overview">{{.Overview}}</td>
<td class="numdocs">{{.NumFilesX}}</td>
<td class="review_date center">{{if .Single}}{{if .Date}}<a class="lookuplink" title="Search for {{.DateYYMM}}" href="find?` + Param_Labels["find"] + `={{.DateYYMM}}&` + Param_Labels["field"] + `=review_date">{{end}}{{end}}{{.ShowDate}}{{if .Single}}{{if .Date}}</a>{{end}}{{end}}</td>
</tr>
`
	templateCreateNewBox = `
<tr>
<td colspan="6"><input type="text" autofocus placeholder="` + prefs.Literals["newboxnumber"] + `" class="keyinput boxid" oninput="checkNewBoxid(this);">
 <input type="button" disabled class="btn" value="` + prefs.Literals["createnewbox"] + `" onclick="startNewBox(this);"></td>
</tr>
`

	templateNewBoxContentLine = `
<tr>
<td><input type="text" autofocus style="width:95%" list="ownerlist" class="keyinput" oninput="newContentSaveNeeded(this);"></td>
<td><input type="text" style="width:95%" list="clientlist" class="keyinput" oninput="fetchClientNamelist(this);newContentSaveNeeded(this);"></td>
<td><input type="text" style="width:95%" list="namelist" oninput="newContentSaveNeeded(this);"></td>
<td><input type="text" style="width:95%" oninput="newContentSaveNeeded(this);"></td>
<td class="date">
#DATESELECTORS#
</td>
<td class="center"><input type="button" class="btn" data-boxid="{{.Boxid}}" disabled value="Add!" onclick="addNewBoxContent(this);">
</tr>
`

}

func initOwnerTemplates() {

	templateOwnerListHead = `
	<table class="ownerlist">
	<thead>
	<tr>
	
	
	<th class="owner">{{if .Single}}{{else}}<a title="&#8645;" class="sortlink" href="/owners?` + Param_Labels["order"] + `=owner{{if .Desc}}&` + Param_Labels["desc"] + `=owner{{end}}">{{end}}` + prefs.Field_Labels["owner"] + `{{if .Single}}{{else}}</a>{{end}}</th>

	<th class="name">{{if .Single}}{{else}}<a title="&#8645;" class="sortlink" href="/owners?` + Param_Labels["order"] + `=name{{if .Desc}}&` + Param_Labels["desc"] + `=name{{end}}">{{end}}` + prefs.Field_Labels["name"] + `{{if .Single}}{{else}}</a>{{end}}</th>
	
	{{if .UpdateOK}}<th></th>{{end}}

	<th class="number">{{if .Single}}{{else}}<a title="&#8645;" class="sortlink" href="/owners?` + Param_Labels["order"] + `=numdocs{{if .Desc}}&` + Param_Labels["desc"] + `=numdocs{{end}}">{{end}}` + prefs.Field_Labels["numdocs"] + `{{if .Single}}{{else}}</a>{{end}}</th>
	</tr>
	</thead>
	<tbody>
	`

	templateOwnerFilesHead = `
	<table class="ownerfiles">
	<thead>
	<tr>
	
	<th class="owner"><a title="&#8645;" class="sortlink" href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=boxid{{if .Desc}}&` + Param_Labels["desc"] + `=boxid{{end}}">` + prefs.Field_Labels["boxid"] + `</a></th>
	<th class="client"><a title="&#8645;" class="sortlink" href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=client{{if .Desc}}&` + Param_Labels["desc"] + `=client{{end}}">` + prefs.Field_Labels["client"] + `</a></th>
	<th class="name"><a title="&#8645;" class="sortlink" href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=name{{if .Desc}}&` + Param_Labels["desc"] + `=name{{end}}">` + prefs.Field_Labels["name"] + `</a></th>
	<th class="contents"><a title="&#8645;" class="sortlink" href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=contents{{if .Desc}}&` + Param_Labels["desc"] + `=contents{{end}}">` + prefs.Field_Labels["contents"] + `</a></th>
	<th class="review_date"><a title="&#8645;" class="sortlink" href="/owners?` + Param_Labels["owner"] + `={{.Owner}}&` + Param_Labels["order"] + `=review_date{{if .Desc}}&` + Param_Labels["desc"] + `=review_date{{end}}">` + prefs.Field_Labels["review_date"] + `</a></th>

	</tr>
	</thead>
	<tbody>
	`

	templateOwnerListLine = `
	<tr>
	<td class="owner">{{if .Single}}{{else}}{{if .Owner}}<a href="/owners?` + Param_Labels["owner"] + `={{.OwnerUrl}}">{{end}}{{end}}{{.Owner}}{{if .Single}}{{else}}{{if .Owner}}</a>{{end}}{{end}}</td>

	<td class="name">
		{{if and .Single .UpdateOK}}
			<input type="hidden" id="AutosaveSeconds" value="` + strconv.Itoa(prefs.AutosaveSeconds) + `">
			<input id="ownername" data-owner="{{.Owner}}" oninput="autosave_OwnerName(this);"  
			value="{{if .Name}}{{.Name}}{{end}}">
		{{else}}
			{{if .Name}}{{.Name}}{{end}}
		{{end}}
	</td>

	{{if and .Single .UpdateOK}}
		<td>
			<input type="button" class="btn hide" id="saveOwnerName" value="Save changes" onclick="updateOwnerName(this);">
			{{if .DeleteOK}}
				<input type="checkbox" title="Enable delete button" onchange="this.nextElementSibling.classList.remove('hide');this.classList.add('hide');">
				<input type="button" class="btn hide" data-owner="{{.Owner}}" value="Delete" onclick="deleteChildlessOwner(this);">
			{{end}}
		</td>
	{{end}}
	<td class="vdata">{{.NumFilesX}}</td>
	</tr>
	`

	templateOwnerFilesLine = `
	<tr>
	<td class="boxid" title="{{.Overview}}">{{if .Boxid}}<a href="/boxes?` + Param_Labels["boxid"] + `={{.BoxidUrl}}">{{end}}{{.Boxid}}{{if .Boxid}}</a>{{end}}</td>
	<td class="client">{{if .Client}}<a href="/find?` + Param_Labels["find"] + `={{.ClientUrl}}&` + Param_Labels["field"] + `=client">{{end}}{{.Client}}{{if .Client}}</a>{{end}}</td>
	<td class="name">{{.Name}}</td>
	<td class="contents">{{.Contents}}</td>
	<td class="review_date center">{{if .Date}}<a class="lookuplink" title="Search for {{.DateYYMM}}" href="/find?` + Param_Labels["find"] + `={{.DateYYMM}}&` + Param_Labels["field"] + `=review_date">{{end}}{{.ShowDate}}{{if .Date}}</a>{{end}}</td>
	
	</tr>
	`

}

func initLocationTemplates() {

	// Header for box listing by location
	templateLocationBoxTableHead = `
		<table class="boxlist">
		<thead>
		<tr>
		<th class="boxid"><a title="&#8645;" class="sortlink" href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}&` + Param_Labels["order"] + `=boxid{{if .Desc}}&` + Param_Labels["desc"] + `=boxid{{end}}">` + prefs.Field_Labels["boxid"] + `</a></th>
		<th class="storeref"><a title="&#8645;" class="sortlink" href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}&` + Param_Labels["order"] + `=storeref{{if .Desc}}&` + Param_Labels["desc"] + `=storeref{{end}}">` + prefs.Field_Labels["storeref"] + `</a></th>
		<th class="contents"><a title="&#8645;" class="sortlink" href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}&` + Param_Labels["order"] + `=overview{{if .Desc}}&` + Param_Labels["desc"] + `=overview{{end}}">` + prefs.Field_Labels["overview"] + `</a></th>
		<th class="boxid"><a title="&#8645;" class="sortlink" href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}&` + Param_Labels["order"] + `=numdocs{{if .Desc}}&` + Param_Labels["desc"] + `=numdocs{{end}}">` + prefs.Field_Labels["numdocs"] + `</a></th>
		<th class="boxid"><a title="&#8645;" class="sortlink" href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}&` + Param_Labels["order"] + `=min_review_date{{if .Desc}}&` + Param_Labels["desc"] + `=min_review_date{{end}}">` + prefs.Field_Labels["review_date"] + `</a></th>
		</tr>
		</thead>
		<tbody>
		`

	templateLocationListHead = `
		<table class="locationlist">
		<thead>
		<tr>
		
		
		<th class="location">{{if .Single}}{{else}}<a title="&#8645;" class="sortlink" href="/locations?` + Param_Labels["order"] + `=location{{if .Desc}}&` + Param_Labels["desc"] + `=location{{end}}">{{end}}` + prefs.Field_Labels["location"] + `{{if .Single}}{{else}}</a>{{end}}</th>
		<th class="numboxes">{{if .Single}}{{else}}<a title="&#8645;" class="sortlink" href="/locations?` + Param_Labels["order"] + `=NumBoxes{{if .Desc}}&` + Param_Labels["desc"] + `=NumBoxes{{end}}">{{end}}` + prefs.Field_Labels["numboxes"] + `{{if .Single}}{{else}}</a>{{end}}</th>
		</tr>
		</thead>
		<tbody>
		`
	templateNewLocation = `
		<tr><td class="location"><input type="text" style="width:95%;" autofocus></td>
		<td><input type="button" class="btn" value="Add new ` + prefs.Field_Labels["location"] + `"
		onclick="addNewLocation(this);"></td></tr>
		`

	templateLocationListLine = `
		<tr>
		<td class="location">{{if .Single}}{{else}}{{if .Location}}<a href="/locations?` + Param_Labels["location"] + `={{.LocationUrl}}">{{end}}{{end}}{{.Location}}{{if .Single}}{{else}}{{if .Location}}</a>{{end}}{{end}}</td>
		<td class="numboxes">{{if .DeleteOK}}<input type="button" class="btn" value="Delete" onclick="deleteLocation(this);">{{else}}{{.NumBoxesX}}{{end}}</td>
		</tr>
		`
	// Header for box listing by location
	templateLocationBoxTableRow = `
	<tr>
	<td class="boxid">{{if .Boxid}}<a href="/boxes?` + Param_Labels["boxid"] + `={{.BoxidUrl}}">{{end}}{{.Boxid}}{{if .Boxid}}</a>{{end}}</td>
	<td class="storeref">{{if .Storeref}}<a class="lookuplink" title="Search for {{.Storeref}}" href="/find?` + Param_Labels["find"] + `={{.StorerefUrl}}&` + Param_Labels["field"] + `=storeref">{{end}}{{.Storeref}}{{if .Storeref}}</a>{{end}}</td>
	<td class="overview">{{.Contents}}</td>
	<td class="numdocs">{{.NumFiles}}</td>
	<td class="review_date center">{{if .Single}}{{if .Date}}<a class="lookuplink" title="Search for {{.DateYYMM}}" href="find?` + Param_Labels["find"] + `={{.DateYYMM}}&` + Param_Labels["field"] + `=review_date">{{end}}{{end}}{{.ShowDate}}{{if .Single}}{{if .Date}}</a>{{end}}{{end}}</td>
	</tr>
	`

}

func initSearchTemplates() {

	templateSearchHome = `
	<p>I'm currently minding <strong>{{.NumDocsX}}</strong> individual files packed into
	<strong>{{.NumBoxesX}}</strong> boxes stored in <strong>{{.NumLocnsX}}
	</strong> locations.</p>
	
	<form action="/find" onsubmit="if (this.fld.value=='') { this.fld.name=''; }">
	<main>You can search the archives by entering the text you're looking for
	here <input type="text" title="Text to find" autofocus name="` + Param_Labels["find"] + `"/>
	<details title="Fields to search" style="display:inline;">
	<summary><strong>&#8799;</strong></summary>
	<select title="Field to search" id="fld" name="` + Param_Labels["field"] + `">
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
	<input type="submit" class="btn" value="Find it!"/><br />
	You can enter a partner's initials, a client number or name, a box number or storage reference, a common term such as <em>tax</em> or a review date* or year.<br>
	Just enter the terms you're looking for, no quote marks, ANDs, ORs, etc.<br>
	* Enter review dates as <em>yyyy</em> or <em>yyyy-mm</em> eg: '2026-03'.
	</main></form>
	<p>If you want to restrict the range of records searched, <a href="/params">specify search options here</a>.</p>
	<p>{{if or .Locations .Owners .ExcludeBeforeYear }}Current search restrictions:- {{if .Locations}}<strong>` + prefs.Field_Labels["location"] + `: {{.Locations}};</strong> {{end}} {{if .Owners}}<strong>` + prefs.Field_Labels["owner"] + `: {{.Owners}};</strong> {{end}} {{if .ExcludeBeforeYear}}<strong>` + prefs.Field_Labels["review_date"] + ` &ge; {{.ExcludeBeforeYear}}</strong>{{end}}{{end}}</p>
	`

	// Search response header before page links
	templateSearchResultsHdr1 = `
<p>
{{if .Find}}I was looking for <span class="searchedfor">{{.Find}}
{{if .Field}} in {{.Field}}{{end}}
{{if .Locations}} [` + prefs.Field_Labels["location"] + `: {{.Locations}}]{{end}}
{{if .Owners}} [` + prefs.Field_Labels["owner"] + `: {{.Owners}}]{{end}}
{{if .ExcludeBeforeYear}} [` + prefs.Field_Labels["review_date"] + `: &ge; {{.ExcludeBeforeYear}}]{{end}}
</span> and
{{else}}{{if or .Locations .Owners}}
	<span class="searchedfor">
	{{if .Locations}} [` + prefs.Field_Labels["location"] + `: {{.Locations}}]{{end}}
	{{if .Owners}} [` + prefs.Field_Labels["owner"] + `: {{.Owners}}]{{end}}
	</span>
{{end}}
{{end}}

I found {{if .Found0}}nothing, nada, rien, zilch.{{end}}{{if .Found1}}just the one match.{{end}}{{if .Found2}}{{.Found}} matches.{{end}}</p>
`

	// Search response header after page links
	templateSearchResultsHdr2 = `
<table class="searchresults">
<thead>
<tr>
<th class="ourbox"><a class="sortlink" href="/find?` + Param_Labels["find"] + `={{.FindUrl}}{{if .OneField}}&` + Param_Labels["field"] + `={{.OneField}}{{end}}&` + Param_Labels["order"] + `=boxid{{if .Desc}}&` + Param_Labels["desc"] + `=boxid{{end}}">` + prefs.Field_Labels["boxid"] + `</a></th>
<th class="owner"><a class="sortlink" href="/find?` + Param_Labels["find"] + `={{.FindUrl}}{{if .OneField}}&` + Param_Labels["field"] + `={{.OneField}}{{end}}&` + Param_Labels["order"] + `=owner{{if .Desc}}&` + Param_Labels["desc"] + `=owner{{end}}">` + prefs.Field_Labels["owner"] + `</a></th>
<th class="client"><a class="sortlink" href="/find?` + Param_Labels["find"] + `={{.FindUrl}}{{if .OneField}}&` + Param_Labels["field"] + `={{.OneField}}{{end}}&` + Param_Labels["order"] + `=client{{if .Desc}}&` + Param_Labels["desc"] + `=client{{end}}">` + prefs.Field_Labels["client"] + `</a></th>
<th class="name"><a class="sortlink" href="/find?` + Param_Labels["find"] + `={{.FindUrl}}{{if .OneField}}&` + Param_Labels["field"] + `={{.OneField}}{{end}}&` + Param_Labels["order"] + `=name{{if .Desc}}&` + Param_Labels["desc"] + `=name{{end}}">` + prefs.Field_Labels["name"] + `</a></th>
<th class="contents"><a class="sortlink" href="/find?` + Param_Labels["find"] + `={{.FindUrl}}{{if .OneField}}&` + Param_Labels["field"] + `={{.OneField}}{{end}}&` + Param_Labels["order"] + `=contents{{if .Desc}}&` + Param_Labels["desc"] + `=contents{{end}}">` + prefs.Field_Labels["contents"] + `</a></th>
<th class="date"><a class="sortlink" href="/find?` + Param_Labels["find"] + `={{.FindUrl}}{{if .OneField}}&` + Param_Labels["field"] + `={{.OneField}}{{end}}&` + Param_Labels["order"] + `=review_date{{if .Desc}}&` + Param_Labels["desc"] + `=review_date{{end}}">` + prefs.Field_Labels["review_date"] + `</a></th>
</tr>
</thead>
<tbody>
`

	templateSearchResultsLine = `
<tr>
<td class="ourbox" title="{{.Overview}}">{{if .Boxid}}<a href="/boxes?` + Param_Labels["boxid"] + `={{.BoxidUrl}}">{{end}}{{.Boxid}}{{if .Boxid}}</a>{{end}}</td>
<td class="owner">{{if .Owner}}<a href="/owners?` + Param_Labels["owner"] + `={{.OwnerUrl}}">{{end}}{{.Owner}}{{if .Owner}}</a>{{end}}</td>
<td class="client">{{if .Client}}<a href="/find?` + Param_Labels["find"] + `={{.ClientUrl}}&` + Param_Labels["field"] + `=client">{{end}}{{.Client}}{{if .Client}}</a>{{end}}</td>
<td class="name">{{.Name}}</td>
<td class="contents">{{.Contents}}</td>
<td class="date center">{{if .Date}}<a class="lookuplink" title="Search for {{.DateYYMM}}" href="/find?` + Param_Labels["find"] + `={{.DateYYMM}}&` + Param_Labels["field"] + `=review_date">{{end}}{{.ShowDate}}{{if .Date}}</a>{{end}}</td>
</tr>
`

	templateSearchParamsHead = `

<form action="/params">
<main>
<p>The settings you choose here will be used to restrict searches until you reset them or until your session ends.</p>
<p><input data-triggered="0" id="savesettings" class="btn" name="` + Param_Labels["savesettings"] + `" disabled onclick="this.setAttribute('data-triggered','1');this.value=this.name;return true;" type="submit" value="Save settings"></p>`

	templateSearchParamsLocationRadios = `
	<div id="locationfilter"><h2>` + prefs.Field_Labels["location"] + `s</h2>
<p>
<input type="radio" id="range_all" name="l` + Param_Labels["range"] + `" value="` + Param_Labels["all"] + `" {{if eq .Lrange "` + Param_Labels["all"] + `"}}checked{{end}} onclick="param_selectLocations(this.checked);">
<label for="range_all"> All </label> &nbsp;&nbsp;&nbsp; 
<input type="radio" id="range_sel" name="l` + Param_Labels["range"] + `" value="` + Param_Labels["selected"] + `" {{if ne .Lrange "` + Param_Labels["all"] + `"}}checked{{end}} onclick="param_selectLocations(!this.checked);">
<label for="range_sel"> Selected only </label>&nbsp;&nbsp;&nbsp;
</p>
`

	templateSearchParamsOwnerRadios = `
	<div id="ownerfilter"><h2>` + prefs.Field_Labels["owner"] + `s</h2>
<p>
<input type="radio" id="orange_all" name="o` + Param_Labels["range"] + `" value="` + Param_Labels["all"] + `" {{if eq .Orange "` + Param_Labels["all"] + `"}}checked{{end}} onclick="param_selectOwners(this.checked);" >
<label for="orange_all"> All </label> &nbsp;&nbsp;&nbsp; 
<input type="radio" id="orange_sel" name="o` + Param_Labels["range"] + `" value="` + Param_Labels["selected"] + `" {{if ne .Orange "` + Param_Labels["all"] + `"}}checked{{end}} onclick="param_selectOwners(!this.checked);">
<label for="orange_sel"> Selected only </label>&nbsp;&nbsp;&nbsp;
</p>
`

	templateSearchParamsDateRadios = `
<div id="datesfilter"><h2>` + prefs.Field_Labels["dates"] + `</h2>
<p>
<input type="radio" id="drange_all" name="d` + Param_Labels["range"] + `" value="` + Param_Labels["all"] + `" {{if eq .Drange "` + Param_Labels["all"] + `"}}checked{{end}} onclick="param_selectDates(this.checked);">
<label for="drange_all"> All </label> &nbsp;&nbsp;&nbsp;&nbsp;
<input type="radio" id="drange_sel" name="d` + Param_Labels["range"] + `" value="` + Param_Labels["selected"] + `"
{{if ne .Drange "` + Param_Labels["all"] + `"}}checked{{end}} onclick="param_selectDates(!this.checked);">
<label for="drange_sel"> Exclude old records </label> &nbsp;&nbsp;&nbsp;&nbsp;
</p>
<p id="daterangedetails"{{if eq .Drange "` + Param_Labels["all"] + `"}} class="hide"{{end}}>
<label for="excludeBeforeYear">Exclude everything before</label>
<input type="number" id="excludeBeforeYear" name="` + Param_Labels["ExcludeBeforeYear"] + `" min="0" max="{{.MaxYear}}" class="year" onchange="enableSaveSettings();" value="{{.ExcludeBeforeYear}}">
</p>
`

}

func initUserTemplates() {

	templateUserLoginHome = `
	<main>
	<h2>Authentication required</h2>
	<form action="/login" method="post">
	<label for="userid">` + prefs.Field_Labels["userid"] + ` </label>
	<input type="text" autofocus id="userid" name="` + Param_Labels["userid"] + `">
	<label for="userpass">` + prefs.Field_Labels["userpass"] + ` </label>
	<input type="password" id="userpass" name="` + Param_Labels["userpass"] + `">
	<input type="submit" value="Authenticate!">
	</form>
	</main>
	`
	templateUserPasswordChange = `
	<p>You may alter your own password by entering the existing password and a new one twice. If you don't know your existing password you'll have to get someone with an accesslevel of ` + prefs.Accesslevels[ACCESSLEVEL_SUPER] + ` to change it for you.</p>
	<form action="/users" method="post" onsubmit="return pwd_validateSingleChange(this);">
	<input type="hidden" name="` + Param_Labels["passchg"] + `" value="` + Param_Labels["single"] + `"|>
	<label for="oldpass">Current password </label> <input autofocus type="password" id="oldpass" name="` + Param_Labels["oldpass"] + `">
	<label for="mynewpass">New password </label> <input type="password" id="mynewpass" name="` + Param_Labels["newpass"] + `">
	<label for="mynewpass2">and again </label> <input type="password" id="mynewpass2">
	<input type="submit" value="Change my password!">
	</form>
	`

	templateMultiUserPasswordChangeHead = `
	<form>
	<input type="hidden" name="` + Param_Labels["passchg"] + `" value="` + Param_Labels["multiple"] + `"|>
	<table id="tabusers">
	<thead><tr>
	<th>Userid</th>
	<th>Accesslevel</th>
	<th>New password</th>
	<th>and again</th>
	<th></th>
	<th></th>
	</tr></thead>
	<tbody>
	`

	templateMultiUserPasswordChangeLine = `
	<tr>
	<td><input title="Userid" type="text" readonly name="m` + Param_Labels["userid"] + `_{{.Row}}" value="{{.Userid}}"></td>
	<td>
		<select title="Accesslevel" name="m` + Param_Labels["accesslevel"] + `_{{.Row}}" onchange="pwd_updateAccesslevel(this);">

		<option value="` + strconv.Itoa(ACCESSLEVEL_UPDATE) + `"{{if eq .Accesslevel ` + strconv.Itoa(ACCESSLEVEL_UPDATE) + `}} selected{{end}}>` + prefs.Accesslevels[ACCESSLEVEL_UPDATE] + `</option>
		<option value="` + strconv.Itoa(ACCESSLEVEL_SUPER) + `"{{if eq .Accesslevel  ` + strconv.Itoa(ACCESSLEVEL_SUPER) + `}} selected{{end}}>` + prefs.Accesslevels[ACCESSLEVEL_SUPER] + `</option>
		</select>
	</td>
	<td><input title="New password" type="password" name="m` + Param_Labels["newpass"] + `_{{.Row}}" oninput="pwd_enableSave(this);"></td>
	<td><input title="New password repeated" type="password" id="newpass2:{{.Row}}" oninput="pwd_enableSave(this);"></td>
	<td><input type="button" disabled value="Set password" onclick="pwd_savePasswordChanges(this);"></td>
	<td class="center">
		<input type="checkbox" title="Enable delete button" name="m` + Param_Labels["deleteuser"] + `_{{.Row}}" value="` + Param_Labels["deleteuser"] + `" onchange="this.parentElement.children[1].disabled=!this.checked;"> 
		<input type="button" disabled value="Delete user" onclick="pwd_deleteUser(this);">
	</td>
	</tr>
	`
	templateMultiUserPasswordChangeFoot = `

	</tbody>
	</table>
	<input type="hidden" id="rowcount" name="` + Param_Labels["rowcount"] + `" value="{{.Row}}">
	<input type="button" value="+" onclick="pwd_insertNewRow(); return false;">
	</form>
	<table class="hide">
	<tr id="newrow" data-fields="newuserid,newal,newpass1,newpass2,savenewuser">
	<td><input type="text" name="m` + Param_Labels["userid"] + `" data-ok="0" oninput="pwd_useridChanged(this);"></td>
	<td>
		<select name="m` + Param_Labels["accesslevel"] + `">

		<option value="` + strconv.Itoa(ACCESSLEVEL_UPDATE) + `">` + prefs.Accesslevels[ACCESSLEVEL_UPDATE] + `</option>
		<option value="` + strconv.Itoa(ACCESSLEVEL_SUPER) + `">` + prefs.Accesslevels[ACCESSLEVEL_SUPER] + `</option>
		</select>
	</td>
	<td><input title="New password" type="password" data-ok="0" name="m` + Param_Labels["newpass"] + `" oninput="pwd_checkpass(this);"></td>
	<td><input title="New password repeated" type="password" data-ok="0" oninput="pwd_checkpass(this);"></td>
	<td><input type="button" disabled value="Save user" onclick="pwd_insertNewUser(this);"></td>
	<td class="hide">
		<input type="checkbox" title="Enable delete button" name="m` + Param_Labels["deleteuser"] + `" value="` + Param_Labels["deleteuser"] + `" onchange="this.parentElement.children[1].disabled=!this.checked;"> 
		<input type="button" disabled value="Delete user" onclick="pwd_deleteUser(this);">
	</td>

	</tr>
	</table>
	`

} // initUserTemplates

func emitTrailer(w http.ResponseWriter) {

	fmt.Fprint(w, `</body></html>`)
}

func emitRootCSS(r *http.Request) string {
	theme := sessionTheme(r)
	res := ":root {\n"
	v := reflect.ValueOf(prefs.Themes[theme])
	f := v.Type()
	for i := 0; i < v.NumField(); i++ {
		x := strings.ToLower(strings.ReplaceAll(f.Field(i).Name, "_", "-"))
		res += fmt.Sprintf("--%v: %v;\n", x, v.Field(i).Interface())
	}
	res += "}\n"
	return res
}

//go:embed normalize.css
var cssreset string

//go:embed embedded.css
var css string

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

type ownerlistvars struct {
	Owner     string
	OwnerUrl  string
	Name      string
	NumFiles  int
	NumFilesX string
	Desc      bool
	NumOrder  bool
	Single    bool
	UpdateOK  bool
	DeleteOK  bool
}

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
	ShowDate  string
	DateYYMM  string
	Overview  string
	Desc      bool
}

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
	ShowDate        string
	DateYYMM        string
	Desc            bool
	Single          bool
	UpdateOK        bool
	DeleteOK        bool
}

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
	ShowDate  string
	DateYYMM  string
	Desc      bool
	DeleteOK  bool
	UpdateOK  bool
	Id        int
	Select    string
}

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

func offerThemesList(theme string) string {

	sel := `<select onchange="setTheme(this.value);">`
	for t := range prefs.Themes {
		sel += `<option value="` + t + `"`
		if theme == t {
			sel += ` selected`
		}
		sel += `> ` + t + ` </option>`
	}
	sel += `</select>`
	return sel

}

func start_html(w http.ResponseWriter, r *http.Request) {

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

	var html2 = `
-->
</style>
</head>
<body onload="bodyLoaded();">
<h1><a href="/">{{.Apptitle}}</a> <span class="themepick" title="Choose colours">###</span> {{if .Updating}} <span style="font-size: 1.2em;" title="Running in Update Mode"> &#9997; </span>{{end}}</h1>
<div class="topmenu"><div class="menulinks">
`

	const errormsgdiv = `<div id="errormsgdiv" class="hide"></div>`

	var basicMenu = `
	<a href="/search">` + prefs.Menu_Labels["search"] + `</a> 
	<a href="/locations">` + prefs.Menu_Labels["locations"] + `</a> 
	<a href="/owners">` + prefs.Menu_Labels["owners"] + `</a> 
	<a href="/boxes">` + prefs.Menu_Labels["boxes"] + `</a>  
	<a href="/update">` + prefs.Menu_Labels["update"] + `</a>   
	<a href="/about">` + prefs.Menu_Labels["about"] + `</a>  
	
	`

	var updateMenu = `
	<a href="/search">` + prefs.Menu_Labels["search"] + `</a> 
	<a href="/locations">` + prefs.Menu_Labels["locations"] + `</a> 
	<a href="/owners">` + prefs.Menu_Labels["owners"] + `</a> 
	<a href="/boxes">` + prefs.Menu_Labels["boxes"] + `</a> 
	<a href="/users">` + prefs.Menu_Labels["users"] + `</a>
	<a href="/logout">` + prefs.Menu_Labels["logout"] + ` {{.Userid}}</a> 
	<a href="/about">` + prefs.Menu_Labels["about"] + `</a> 
	`

	var ht string

	updating, usr, _ := updateok(r)
	runvars.Updating = updating

	ht = html1 + cssreset
	ht += emitRootCSS(r)

	ht += css + strings.ReplaceAll(html2, "###", offerThemesList(sessionTheme(r)))

	if !runvars.Updating {
		ht += mark_current_menu_path(basicMenu, r.URL.Path)
	} else {
		if usr != nil {
			runvars.Userid = usr.(string)
		} else {
			runvars.Userid = ""
		}
		ht += mark_current_menu_path(updateMenu, r.URL.Path)
	}
	ht += "</div></div>" + errormsgdiv
	html, err := template.New("mainmenu").Parse(ht)
	checkerr(err)

	html.Execute(w, runvars)

}

func mark_current_menu_path(menu string, path string) string {

	res := strings.Replace(menu, `href="`+path, `class="currentlink" href="`+path, 1)
	return res
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

	fmt.Fprintf(w, `<div class="pagelinks"><span id="pagelinks">`)
	thisPage := (offset / pagesize) + 1
	if thisPage > 1 {
		prevPageOffset := (thisPage * pagesize) - (2 * pagesize)
		fmt.Fprintf(w, `<span class="pagelink"><a id="prevpage" href="/%v?%v`+Param_Labels["offset"]+`=%v" title="Previous page">%v</a></span>`, cmd, varx, prevPageOffset, ArrowPrevPage)
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
				fmt.Fprintf(w, `<span class="pagelink">&nbsp; <strong>%v</strong> &nbsp;</span>`, thisPage)
			} else {
				pOffset := (pageNum * pagesize) - pagesize

				fmt.Fprintf(w, `<span class="pagelink"><a href="/%v?%v`+Param_Labels["offset"]+`=%v" title="">%v</a></span>`, cmd, varx, pOffset, strconv.Itoa(pageNum))
			}
		} else if pageNum == thisPage-(prefs.MaxAdjacentPagelinks+1) || pageNum == thisPage+prefs.MaxAdjacentPagelinks+1 {
			fmt.Fprintf(w, " ... ")
		}
	}
	if thisPage < numPages {
		nextPageOffset := (thisPage * pagesize)
		fmt.Fprintf(w, `<span class="pagelink"><a id="nextpage" href="/%v?%v`+Param_Labels["offset"]+`=%v" title="Next page">%v</a></span>`, cmd, varx, nextPageOffset, ArrowNextPage)
	}

	fmt.Fprint(w, `<span class="pagelink"><select title="Choose results page size" onchange="changePagesize(this);"></span>`)
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

	fmt.Fprintf(w, `</span></div>`)

	return res
}

func ajax_setPagesize(w http.ResponseWriter, r *http.Request) {

	pagesize, err := strconv.Atoi(r.FormValue(Param_Labels["pagesize"]))
	if err != nil {
		pagesize = prefs.DefaultPagesize
	}
	setPagesize(w, r, pagesize)

	fmt.Fprint(w, `{"res":"ok"}`)

}
func ajax_setTheme(w http.ResponseWriter, r *http.Request) {

	theme := r.FormValue(Param_Labels["theme"])
	if theme == "" {
		return
	}

	printDebug("Setting theme = " + theme)
	setTheme(w, r, theme)

	fmt.Fprint(w, `{"res":"ok"}`)

}

type vars struct {
	Regular_background   string
	Regular_foreground   string
	Hilite_background    string
	Hilite_foreground    string
	Link_color           string
	Link_hilite_back     string
	Link_hilite_fore     string
	Button_background    string
	Button_foreground    string
	Disabled_background  string
	Disabled_foreground  string
	Cell_background      string
	Cell_border_color    string
	Pagelinks_background string
	Edit_background      string
	Edit_foreground      string
	Error_background     string
	Error_foreground     string
}
type userpreferences struct {
	HttpPort             string            `yaml:"httpPort"`
	MaxAdjacentPagelinks int               `yaml:"MaxAdjacentPagelinks"`
	Accesslevels         map[int]string    `yaml:"AccesslevelNames"`
	MaxBoxContents       int               `yaml:"MaxBoxContents"`
	Field_Labels         map[string]string `yaml:"FieldLabels"`
	Menu_Labels          map[string]string `yaml:"MenuLabels"`
	Table_Labels         map[string]string `yaml:"TableLabels"`
	Literals             map[string]string `yaml:"Literals"`
	HistoryLog           map[string]int    `yaml:"HistoryLog"`
	AppTitle             string            `yaml:"AppTitle"`
	CookieMaxAgeMins     int               `yaml:"LoginMinutes"`
	PasswordMinLength    int               `yaml:"PasswordMinLength"`
	DefaultPagesize      int               `yaml:"DefaultPagesize"`
	Pagesizes            []int             `yaml:"PagesizeOptions"`
	FixLazyTyping        []string          `yaml:"FixAllLowercaseFields"`
	FuturePicklistYears  int               `yaml:"FuturePicklistYears"`
	AutosaveSeconds      int               `yaml:"AutosaveSeconds"`
	DefaultReviewMonths  int               `yaml:"DefaultReviewMonths"`
	ShowDateFormat       string            `yaml:"ShowDateFormat"`
	IncludePastYears     int               `yaml:"IncludePastYears"`
	DefaultTheme         string            `yaml:"DefaultTheme"`
	Themes               map[string]vars   `yaml:"Themes"`
}

// YAML format configuration
//
//go:embed embedded.yaml
var internal_config string
