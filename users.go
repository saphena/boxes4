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

// ajax handler
func insertNewUser(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	checkerr(err)

	al := r.FormValue(Param_Labels["accesslevel"])
	uid := r.FormValue(Param_Labels["adduser"])
	np1 := r.FormValue(Param_Labels["newpass"])
	np2 := r.FormValue(Param_Labels["newpass2"])

	sqlx := "SELECT userid FROM users WHERE userid LIKE '" + strings.ReplaceAll(uid, "'", "''") + "'"
	fmt.Println("DEBUG: " + sqlx)
	if getValueFromDB(sqlx, "") != "" {
		fmt.Fprintf(w, `{"res":"User %v already exists!"}`, uid)
		return
	}
	if len(np1) < prefs.PasswordMinLength {
		fmt.Fprintf(w, `{"res":"Can't create user %v, password %v is too short %v."}`, uid, len(np1), prefs.PasswordMinLength)
		return
	}
	if np1 != np2 {
		fmt.Fprintf(w, `{"res":"Can't create user %v, the two passwords don't match."}`, uid)
		return
	}
	sqlx = "INSERT INTO users (userid,userpass,accesslevel) VALUES("
	sqlx += "'" + strings.ReplaceAll(uid, "'", "''") + "'"
	sqlx += ",'" + strings.ReplaceAll(np1, "'", "''") + "'"
	sqlx += "," + al + ")"
	res := DBExec(sqlx)
	n, err := res.RowsAffected()
	checkerr(err)
	if n < 1 {
		fmt.Fprint(w, `{"res":"Insert failed!"}`)
		return
	}

	fmt.Fprint(w, `{"res":"ok"}`)
}

func showusers(w http.ResponseWriter, r *http.Request) {

	type MultiPassVars = struct {
		Userid      string
		Accesslevel int
		Row         int
	}
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

		fmt.Fprint(w, templateUserPasswordChange)
		fmt.Fprint(w, "</main>")
		return
	}
	fmt.Fprint(w, templateMultiUserPasswordChangeHead)
	temp, err := template.New("multiUserPasswordChangeLine").Parse(templateMultiUserPasswordChangeLine)
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
	temp, err = template.New("multiUserPasswordChangeFoot").Parse(templateMultiUserPasswordChangeFoot)
	checkerr(err)
	err = temp.Execute(w, v)
	checkerr(err)

	fmt.Fprint(w, "<hr>")
	fmt.Fprint(w, templateUserPasswordChange)

	fmt.Fprint(w, "</main>")
}

func ajax_users(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	if r.FormValue(Param_Labels["adduser"]) != "" {
		insertNewUser(w, r)
		return
	}
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
			pwd2 := r.FormValue(Param_Labels["newpass2"])

			if len(pwd) < prefs.PasswordMinLength {
				fmt.Fprintf(w, `{"res":"Password is too short, min length is %v chars."}`, prefs.PasswordMinLength)
				return
			}
			if pwd != pwd2 {
				fmt.Fprint(w, `{"res":"Passwords don't match"}`)
				return
			}
			if len(pwd) > 1 {
				sqlx = "UPDATE users SET userpass='" + strings.ReplaceAll(pwd, "'", "''") + "' "
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
