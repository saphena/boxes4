package main

import (
	"html/template"
	"net/http"
)

func update(w http.ResponseWriter, r *http.Request) {

	start_html(w, r)

	temp, err := template.New("loginscreen").Parse(loginscreen)
	if err != nil {
		panic(err)
	}

	err = temp.Execute(w, "")
	if err != nil {
		panic(err)
	}
}
