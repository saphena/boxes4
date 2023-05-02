package main

import (
	"html/template"
	"net/http"
)

func update(w http.ResponseWriter, r *http.Request) {

	var loginscreen = `
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

	start_html(w, r)

	temp, err := template.New("loginscreen").Parse(loginscreen)
	checkerr(err)

	err = temp.Execute(w, "")
	checkerr(err)
}
