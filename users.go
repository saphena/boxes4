package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

func updateMultipleUsers(w http.ResponseWriter, r *http.Request) {

	rowcount, _ := strconv.Atoi(r.PostFormValue(Param_Labels["rowcount"]))
	if rowcount < 1 {
		return
	}
	fmt.Printf("DEBUG: updateMultipleUsers = %v\n", rowcount)
	for i := 1; i <= rowcount; i++ {
		sqlx := ""
		ix := strconv.Itoa(i)
		uid := r.PostFormValue("m" + Param_Labels["userid"] + "_" + ix)
		if r.PostFormValue("m"+Param_Labels["deleteuser"]+"_"+ix) == Param_Labels["deleteuser"] {
			sqlx = "DELETE FROM users WHERE userid='" + uid + "'"
		} else {
			pwd := r.PostFormValue("m" + Param_Labels["newpass"] + "_" + ix)
			al := r.PostFormValue("m" + Param_Labels["accesslevel"] + "_" + ix)
			if al == "" {
				fmt.Printf("DEBUG: al is blank for row %v uid='%v'\n", i, uid)
				continue
			}
			sqlx = "UPDATE users SET accesslevel=" + al
			if len(pwd) > 1 {
				sqlx += ", userpass=`" + strings.ReplaceAll(pwd, "'", "''") + "' "
			}
			sqlx += " WHERE userid='" + uid + "'"
		}
		if sqlx != "" {
			fmt.Printf("DEBUG: %v\n", sqlx)
			res := DBExec(sqlx)
			n, err := res.RowsAffected()
			checkerr(err)
			if n < 1 {
				fmt.Fprint(w, "One or more updates failed")
			}
		}
	}
}
func changeSinglePassword(w http.ResponseWriter, r *http.Request) {

	oldpass := strings.ReplaceAll(r.PostFormValue(Param_Labels["oldpass"]), "'", "''")
	newpass := strings.ReplaceAll(r.PostFormValue(Param_Labels["newpass"]), "'", "''")
	sqlx := "UPDATE users SET userpass='" + newpass + "' WHERE userid='" + runvars.Userid + "' AND userpass='" + oldpass + "'"
	res := DBExec(sqlx)
	n, err := res.RowsAffected()
	checkerr(err)
	if n == 1 {
		fmt.Fprint(w, `<p>Password changed successfully.</p`)
	} else {
		fmt.Fprint(w, `<p class="errormsg">Password change unsuccessful.</p>`)
	}

}

func showusers(w http.ResponseWriter, r *http.Request) {

	var userpasschg = `
	<p>You may alter your own password by entering the existing password and a new one twice. If you don't know your existing password you'll have to get someone with an accesslevel of ` + prefs.Accesslevels[ACCESSLEVEL_SUPER] + ` to change it for you.</p>
	<form action="/users" method="post" onsubmit="return pwd_validateSingleChange(this);">
	<input type="hidden" name="` + Param_Labels["passchg"] + `" value="` + Param_Labels["single"] + `"|>
	<input type="hidden" id="minpwlen" value="` + strconv.Itoa(prefs.PasswordMinLength) + `">
	<label for="oldpass">Current password </label> <input autofocus type="password" id="oldpass" name="` + Param_Labels["oldpass"] + `">
	<label for="newpass">New password </label> <input type="password" id="newpass" name="` + Param_Labels["newpass"] + `">
	<label for="newpass2">and again </label> <input type="password" id="newpass2">
	<input type="submit" value="Change my password!">
	</form>
	`
	var multipasschghead = `
	<form action="/users" method="post">
	<input type="hidden" name="` + Param_Labels["passchg"] + `" value="` + Param_Labels["multiple"] + `"|>
	<input type="hidden" id="minpwlen" value="` + strconv.Itoa(prefs.PasswordMinLength) + `">
	<table>
	<thead><tr>
	<th>Userid</th>
	<th>Accesslevel</th>
	<th>New password</th>
	<th>and again</th>
	<th>Delete user</th>
	</tr></thead>
	<tbody>
	`

	type MultiPassVars = struct {
		Userid      string
		Accesslevel int
		Row         int
	}
	var multipassline = `
	<tr>
	<td><input type="text" readonly name="m` + Param_Labels["userid"] + `_{{.Row}}" value="{{.Userid}}"></td>
	<td>
		<select name="m` + Param_Labels["accesslevel"] + `_{{.Row}}" onchange="pwd_deleteUser(this);">
		<option value="` + strconv.Itoa(ACCESSLEVEL_READONLY) + `"{{if eq .Accesslevel ` + strconv.Itoa(ACCESSLEVEL_READONLY) + `}} selected{{end}}>` + prefs.Accesslevels[ACCESSLEVEL_READONLY] + `</option>
		<option value="` + strconv.Itoa(ACCESSLEVEL_UPDATE) + `"{{if eq .Accesslevel ` + strconv.Itoa(ACCESSLEVEL_UPDATE) + `}} selected{{end}}>` + prefs.Accesslevels[ACCESSLEVEL_UPDATE] + `</option>
		<option value="` + strconv.Itoa(ACCESSLEVEL_SUPER) + `"{{if eq .Accesslevel  ` + strconv.Itoa(ACCESSLEVEL_SUPER) + `}} selected{{end}}>` + prefs.Accesslevels[ACCESSLEVEL_SUPER] + `</option>
		</select>
	</td>
	<td><input type="password" name="m` + Param_Labels["newpass"] + `_{{.Row}}"></td>
	<td><input type="password" id="newpass2:{{.Row}}"></td>
	<td class="center"><input type="checkbox" name="m` + Param_Labels["deleteuser"] + `_{{.Row}}" value="` + Param_Labels["deleteuser"] + `"></td>
	</tr>
	`
	var multipasschgfoot = `
	</tbody>
	</table>
	<input type="hidden" name="` + Param_Labels["rowcount"] + `" value="{{.Row}}">
	<input type="submit" value="Update changes">
	</form>
	`
	err := r.ParseForm()
	checkerr(err)

	ok, usr, al := updateok(r)
	if !ok {
		show_search(w, r)
		return
	}
	start_html(w, r)

	//fmt.Fprintf(w, "DEBUG: %v<hr>", r)

	if r.PostFormValue(Param_Labels["passchg"]) == Param_Labels["single"] {
		changeSinglePassword(w, r)
		return
	}
	if r.PostFormValue(Param_Labels["passchg"]) == Param_Labels["multiple"] {
		updateMultipleUsers(w, r)
		//return
	}
	fmt.Fprint(w, "<main>")
	fmt.Fprintf(w, "<p>Hello %v, your accesslevel is %v</p>", usr, prefs.Accesslevels[al.(int)])

	if al.(int) < ACCESSLEVEL_SUPER {

		fmt.Fprint(w, userpasschg)
		fmt.Fprint(w, "</main>")
		return
	}
	fmt.Fprint(w, multipasschghead)
	temp, err := template.New("multipassline").Parse(multipassline)
	checkerr(err)
	sqlx := "SELECT userid,accesslevel FROM users WHERE userid<>'" + runvars.Userid + "' ORDER BY userid"
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	var v MultiPassVars
	for rows.Next() {
		rows.Scan(&v.Userid, &v.Accesslevel)
		v.Row++
		err = temp.Execute(w, v)
		if err != nil {
			panic(err)
		}
	}
	temp, err = template.New("multipasschgfoot").Parse(multipasschgfoot)
	checkerr(err)
	err = temp.Execute(w, v)
	checkerr(err)

	fmt.Fprint(w, "<hr>")
	fmt.Fprint(w, userpasschg)

	fmt.Fprint(w, "</main>")
}

func ajax_users(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	if r.FormValue(Param_Labels["passchg"]) != Param_Labels["single"] {
		fmt.Printf("DEBUG: '%v' / '%v'\n", r.FormValue(Param_Labels["passchg"]), Param_Labels["single"])
		fmt.Fprint(w, `{"res":"NOT IMPLEMENTED"}`)
		return
	}
	uid := r.FormValue(Param_Labels["userid"])
	sqlx := ""
	if r.FormValue(Param_Labels["deleteuser"]) == Param_Labels["deleteuser"] {
		sqlx = "DELETE FROM users "
	} else {
		al := r.FormValue(Param_Labels["accesslevel"])
		if al != "" {
			sqlx = "UPDATE users SET accesslevel=" + al
		} else {
			pwd := r.FormValue(Param_Labels["newpass"])

			if len(pwd) > 1 {
				sqlx = "UPDATE users SET userpass=`" + strings.ReplaceAll(pwd, "'", "''") + "' "
			}
		}
	}
	if sqlx != "" {
		sqlx += " WHERE userid='" + uid + "'"
		fmt.Printf("DEBUG: %v\n", sqlx)
		res := DBExec(sqlx)
		n, err := res.RowsAffected()
		checkerr(err)
		if n < 1 {
			fmt.Fprint(w, `{"res":"One or more updates failed"}`)
			return
		}
	}
	fmt.Fprint(w, `{"res":"ok"}`)

}
