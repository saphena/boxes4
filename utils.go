package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// I insert commas in string representation of integer
func commas(n int) string {
	in := strconv.Itoa(n)
	numOfDigits := len(in)
	if n < 0 {
		numOfDigits-- // First character is the - sign (not a digit)
	}
	numOfCommas := (numOfDigits - 1) / 3

	out := make([]byte, len(in)+numOfCommas)
	if n < 0 {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}

func loadConfiguration(cfgfile *string) {

	file := strings.NewReader(internal_config)
	D := yaml.NewDecoder(file)
	err := D.Decode(&prefs)
	if err != nil {
		panic(err)
	}

	if *cfgfile == "" {
		*cfgfile = "boxes.yaml"
	}
	yml, err := os.ReadFile(*cfgfile)
	if err == nil {

		file = strings.NewReader(string(yml))
		D = yaml.NewDecoder(file)
		err = D.Decode(&prefs)
		if err != nil {
			panic(err)
		}
	}

	runvars = AppVars{prefs.AppTitle, "", script}
	//fmt.Printf("Field_Labels = %v\n", prefs.Field_Labels)
	//fmt.Printf("Port is %v Accesslevels labels - %v\n", prefs.HttpPort, prefs.Accesslevels)
}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

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
