package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

// This is concerned with session variables for logged-in users.

const cookie_name = "boxes4docarchives"

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

func secret(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, cookie_name)

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Print secret message
	fmt.Fprintln(w, "The cake is a lie!")
}

func login(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, cookie_name)

	// Authentication goes here
	// ...

	userid := strings.ReplaceAll(strings.ToLower(r.FormValue(Param_Labels["userid"])), "'", "''")
	password := strings.ToLower(r.FormValue(Param_Labels["userpass"]))
	sqlx := "SELECT userpass,accesslevel FROM users WHERE userid='" + userid + "'"
	var passwd string
	var alevel int
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	if !rows.Next() {
		reject_login(w, r)
		return
	}
	rows.Scan(&passwd, &alevel)
	if passwd != password {
		reject_login(w, r)
		return
	}
	if alevel < ACCESSLEVEL_UPDATE {
		reject_login(w, r)
		return
	}
	// Set user as authenticated
	session.Values["authenticated"] = true
	session.Values["userid"] = r.FormValue(Param_Labels["userid"])
	session.Values["accesslevel"] = alevel
	if session.Values["theme"] == nil {
		session.Values["theme"] = prefs.DefaultTheme
	}
	session.Options.MaxAge = 60 * prefs.CookieMaxAgeMins
	session.Options.HttpOnly = true
	session.Options.SameSite = http.SameSiteStrictMode

	session.Options.Secure = false // we connect using http.
	session.Save(r, w)
	//http.Redirect(w, r, "/search", http.StatusAccepted)
	show_search(w, r)
}

func sessionTheme(r *http.Request) string {

	session, err := store.Get(r, cookie_name)
	checkerr(err)
	printDebug(fmt.Sprintf("sessionTheme is %v\n", session.Values["theme"]))
	if session.Values["theme"] == nil {
		return prefs.DefaultTheme
	}
	return session.Values["theme"].(string)
}

func updateok(r *http.Request) (bool, any, any) {

	session, err := store.Get(r, cookie_name)
	checkerr(err)
	if session.Values["authenticated"] == nil || session.Values["authenticated"].(bool) != true {
		return false, "", ACCESSLEVEL_READONLY
	}
	return true, session.Values["userid"], session.Values["accesslevel"]
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, cookie_name)

	// Revoke users authentication
	session.Values["authenticated"] = false
	session.Values["userid"] = ""
	session.Save(r, w)
	show_search(w, r)
}

func reject_login(w http.ResponseWriter, r *http.Request) {

	start_html(w, r)
	fmt.Fprintf(w, "<p>Login failed!</p>")
}
