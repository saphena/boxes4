package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func about(w http.ResponseWriter, r *http.Request) {

	start_html(w, r)
	fmt.Fprint(w, "<h2>BOXES version 4.0</h2>")
	fmt.Fprint(w, `<p class='copyrite'>Copyright &copy; 2023 Bob Stammers <a href="mailto:stammers.bob@gmail.com">stammers.bob@gmail.com</a> </p>`)

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	updating, usr, alevel := updateok(r)
	lastUpdated := getValueFromDB("SELECT recordedat FROM history ORDER BY recordedat DESC LIMIT 0,1", "recordedat", "")
	if lastUpdated != "" {
		updatedBy := getValueFromDB("SELECT userid FROM history ORDER BY recordedat DESC LIMIT 0,1", "userid", "")
		tsfmt := "2006-01-02T15:04:05Z"
		ts, err := time.Parse(tsfmt, lastUpdated)
		if err != nil {
			fmt.Fprint(w, err)
		}
		fmt.Fprintf(w, "<p>Database last updated <strong>%v</strong> by '%v'</p>", ts.Format("Monday 2 Jan 2006 @ 3:04pm"), updatedBy)
	} else {
		fmt.Fprint(w, "<p>Database not updated</p>")
	}

	if !updating {
		fmt.Fprint(w, `<p>Click [update] above and login as a user with CONTROLLER accesslevel to get more info. `)
		var uids []string
		rows, err := DBH.Query("SELECT userid FROM users WHERE accesslevel >= " + strconv.Itoa(ACCESSLEVEL_UPDATE))
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		for rows.Next() {
			var uid string
			rows.Scan(&uid)
			uids = append(uids, uid)
		}
		fmt.Fprintf(w, ` The following userids have that accesslevel: <strong>%v</strong></p>`, uids)

		return
	}

	fmt.Fprintf(w, `<p>You are logged in as <span class="upper">%v</span> with access level <span class="upper">%v</span></p><hr>`, usr, ACCESSLEVELS[alevel.(int)])

	servername, _ := os.Hostname()

	fmt.Fprintf(w, `<p>I am a server application running on a computer called %v</p>`, servername)
	exPath := filepath.Dir(ex)
	fmt.Fprintf(w, "<p>I'm installed in the folder <strong>%v</strong></p>", exPath)

	type tabledtl struct {
		Table string
		Csvok bool
	}
	tables := []tabledtl{{"boxes", true}, {"contents", true}, {"locations", true}, {"history", false}, {"users", false}}

	fmt.Fprint(w, "<ul>")
	for _, tab := range tables {
		sqlx := "SELECT Count(*) As Rex FROM " + tab.Table
		rex := getValueFromDB(sqlx, "Rex", "0")
		fmt.Fprintf(w, `<li>Table <span class="keydata">%v</span> has <span class="keydata">%v</span> records `, tab.Table, rex)

		if tab.Csvok {
			fmt.Fprintf(w, ` &nbsp;&nbsp;[<a href="/csvexp?%v=%v">Save as CSV</a>]`, Param_Labels["table"], tab.Table)
			fmt.Fprintf(w, ` &nbsp;&nbsp;[<a href="/jsonexp?%v=%v">Save as JSON</a>]`, Param_Labels["table"], tab.Table)
		}

		fmt.Fprint(w, `</li>`)
	}
	fmt.Fprint(w, "</ul><hr>")

	fmt.Fprint(w, `<p><a class="btn" href="/check">Check database</a></p>`)

}
