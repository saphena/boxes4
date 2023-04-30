package main

import (
	"fmt"
	"net/http"
)

func showusers(w http.ResponseWriter, r *http.Request) {

	var userpasschg = `
	<main>
	<p class="errormsg" id="errormsg"></p>
	<p>You may alter your password by entering the existing password and a new one twice. If you don't know your existing password you'll have to get someone with an accesslevel of ` + prefs.Accesslevels[ACCESSLEVEL_SUPER] + ` to change it for you.</p>
	<form action="/users" method="post" onsubmit="document.getElementById('errormsg').innerText='Wrong!';return false;">
	<input type="hidden" name="` + Param_Labels["passchg"] + `" value="` + Param_Labels["single"] + `"|>
	<label for="oldpass">Current password </label> <input autofocus type="password" id="oldpass" name="` + Param_Labels["oldpass"] + `">
	<label for="newpass">New password </label> <input type="password" id="newpass" name="` + Param_Labels["newpass"] + `">
	<label for="newpass2">and again </label> <input type="password" id="newpass2">
	<input type="submit" value="Change my password!">
	</form>
	</main>
	`
	ok, usr, al := updateok(r)
	if !ok {
		show_search(w, r)
		return
	}
	start_html(w, r)
	fmt.Fprintf(w, "<p>Hello %v, your accesslevel is %v</p>", usr, prefs.Accesslevels[al.(int)])
	fmt.Fprint(w, userpasschg)

}
