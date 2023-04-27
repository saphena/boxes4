package main

import (
	"fmt"
	"net/http"
)

func showusers(w http.ResponseWriter, r *http.Request) {

	ok, usr, al := updateok(r)
	if !ok {
		show_search(w, r)
		return
	}
	start_html(w, r)
	fmt.Fprintf(w, "<p>Hello %v, your accesslevel is %v</p>", usr, prefs.Accesslevels[al.(int)])

}
