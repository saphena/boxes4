package main

import (
	"html/template"
	"net/http"
)

func update(w http.ResponseWriter, r *http.Request) {

	start_html(w, r)

	temp, err := template.New("userLoginHome").Parse(templateUserLoginHome)
	checkerr(err)

	err = temp.Execute(w, "")
	checkerr(err)
	emitTrailer(w)
}
