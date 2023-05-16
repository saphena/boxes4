package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	yaml "gopkg.in/yaml.v2"
)

func generateLocationPicklist(loc string, onchange string) string {

	sqlx := "SELECT location FROM locations ORDER BY location"
	rows, err := DBH.Query(sqlx)
	checkerr(err)
	defer rows.Close()
	res := `<select onchange="` + onchange + `">`
	for rows.Next() {
		var locn string
		rows.Scan(&locn)
		res += `<option value="` + locn + `"`
		if loc == locn {
			res += " selected "
		}
		res += ">" + locn + "</option>"
	}
	res += "</select>"
	return res
}

func generateDatePicklist(iso8601dt string, datefieldname string, onchange string) string {

	currentYear := time.Now().Year()
	dataDate := strings.Split(iso8601dt, "-")
	dataYear, _ := strconv.Atoi(dataDate[0])
	dataMonthx := dataDate[1]
	dataMonth, _ := strconv.Atoi(dataMonthx)
	minYear := currentYear
	if dataYear < minYear {
		minYear = dataYear
	}
	maxYear := currentYear + prefs.FuturePicklistYears
	if dataYear > maxYear {
		maxYear = dataYear
	}
	fld := `<input type="hidden" name="` + datefieldname + `" value="` + iso8601dt + `" onchange="` + onchange + `">`
	mths := `<select onchange="date_from_selects(this);">`
	for i := 1; i <= 12; i++ {
		mths += `<option value="` + fmt.Sprintf("%02d", i) + `"`
		if i == dataMonth {
			mths += " selected "
		}
		mths += ">" + time.Month(i).String()
		mths += "</option>"
	}
	mths += "</select>"

	years := `<select onchange="date_from_selects(this);">`
	for i := minYear; i <= maxYear; i++ {
		years += `<option value="` + fmt.Sprintf("%d", i) + `"`
		if i == dataYear {
			years += " selected "
		}
		years += ">" + fmt.Sprintf("%d", i)
		years += "</option>"
	}
	years += "</select>"

	return fld + mths + years

}
func fixAllLowercase(s string) string {

	caser := cases.Title(language.English)
	return caser.String(s)

}

func safesql(s string) string {

	return strings.ReplaceAll(s, "'", "''")

}

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

	runvars = AppVars{prefs.AppTitle, "", script, false}
	//fmt.Printf("Field_Labels = %v\n", prefs.Field_Labels)
	//fmt.Printf("Port is %v Accesslevels labels - %v\n", prefs.HttpPort, prefs.Accesslevels)
}

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
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
	return prefs.DefaultPagesize
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
