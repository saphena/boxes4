package main

import (
	"net/http"
	"strconv"
)

func getValueFromDB(sqlx string, col string, defval string) string {

	rows, err := DBH.Query(sqlx)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var res string
	if !rows.Next() {
		return defval
	}
	rows.Scan(&res)
	return res
}

func rangepagesize(r *http.Request) int {

	n, _ := strconv.Atoi(r.FormValue(Param_Labels["pagesize"]))
	if r.FormValue(Param_Labels["pagesize"]) != "" {
		return n
	}
	return 20
}

func rangeoffset(r *http.Request) int {

	n, _ := strconv.Atoi(r.FormValue(Param_Labels["offset"]))
	return n

}

func order_dir(r *http.Request, field string) string {
	if (r.FormValue(Param_Labels["order"]) != field) || (r.FormValue(Param_Labels["order"]) == r.FormValue(Param_Labels["desc"])) {
		return string("")
	}
	return "&amp;DESC=" + r.FormValue(Param_Labels["order"])
}
