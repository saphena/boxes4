package main

type AppVars struct {
	Apptitle string
	Topmenu  string
}

const css = `
body 				{
	background-color: #FFFFE0;
	font-family:Verdana, Arial;
	font-size: 10pt;
	margin: 1em;
	margin-top: 6px;
	margin-bottom: 6px;
}
div.topmenu a       { text-decoration: none }
a:hover             { text-transform: uppercase; font-weight: bold; }
p 					{ font-family: Verdana, Arial, Helvetica; font-size: 10pt }
p.center			{ text-align: center }
address 			{ font-family: Verdana, Arial, Helvetica; font-size: 8pt }
span				{ font-family: Verdana, Arial, Helvetica; font-size: 10pt; }
table				{}
td 					{ font-family: Verdana, Arial, Helvetica;  padding: 4px; text-align: left }
span.print			{ font-size: 8pt; }

span.required 		{ font-size: 8pt; color: #bb0000; }
span.small 			{ font-size: 8pt }
span.pagetitle		{ font-size: 12pt; font-weight: bold; text-align: center }
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

.topmenu 			{
	display: block;
	border-bottom: solid;
	padding-bottom: 3px;
	margin-bottom: 3px;
	width: 100%;
}
`

const basicMenu = `
[<a href="index.php" accesskey="s">search</a>] 
[<a href="index.php?CMD=SHOWLOCN" accesskey="l">locations</a>] 
[<a href="index.php?CMD=SHOWPTNR" accesskey="p">partners</a>] 
[<a href="index.php?CMD=BOXLIST" accesskey="b">boxes</a>] 
[<a href="index.php?CMD=UPDATE" accesskey="u">update</a>] 
[<a href="index.php?CMD=ABOUT" accesskey="a">about</a>] 

`

const updateMenu = `
[<a href="index.php" accesskey="s">search</a>] 
[<a href="index.php?CMD=SHOWLOCN" accesskey="l">locations</a>] 
[<a href="index.php?CMD=SHOWPTNR" accesskey="p">partners</a>] 
[<a href="index.php?CMD=BOXLIST" accesskey="b">boxes</a>] 
[<a href="index.php?CMD=USERS" accesskey="u">users</a>] 
[<a href="index.php?CMD=LOGOUT" accesskey="l">logout {{.Username}</a>] 
[<a href="index.php?CMD=ABOUT" accesskey="a">about</a>] 

`

const html1 = `
<!DOCTYPE html>
<html lang="en">
<head>
<title>{{.Apptitle}}</title>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
<!--
`
const html2 = `
-->
</style>
</head>
<body>
<h1>{{.Apptitle}}</h1>
<div class="topmenu">
`
