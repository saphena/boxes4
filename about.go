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
	fmt.Fprint(w, "<h2>BOXES4 version 0.1</h2>")
	fmt.Fprint(w, `<p class='copyrite'>Copyright &copy; 2023 Bob Stammers <a href="mailto:stammers.bob@gmail.com">stammers.bob@gmail.com</a> </p>`)

	ex, err := os.Executable()
	checkerr(err)
	updating, usr, alevel := updateok(r)
	lastUpdated := getValueFromDB("SELECT recordedat FROM history ORDER BY recordedat DESC LIMIT 0,1", "recordedat", "")
	if lastUpdated != "" {
		updatedBy := getValueFromDB("SELECT userid FROM history ORDER BY recordedat DESC LIMIT 0,1", "userid", "")
		tsfmt := time.RFC3339 //"2006-01-02T15:04:05Z"
		ts, err := time.Parse(tsfmt, lastUpdated)
		if err != nil {
			fmt.Fprint(w, err)
		}
		fmt.Fprintf(w, "<p>Database last updated <strong>%v</strong> by '%v'</p>", ts.Format("Monday 2 Jan 2006 @ 3:04pm"), updatedBy)
	} else {
		fmt.Fprint(w, "<p>Database not updated</p>")
	}

	if !updating {
		fmt.Fprint(w, `<p>Click [`+prefs.Menu_Labels["update"]+`] above and login as a user with CONTROLLER accesslevel to get more info. `)
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

		fmt.Fprint(w, `<hr><h4>Terms used by this application</h4>`)
		show_terminology(w, r)

		return
	}

	fmt.Fprintf(w, `<p>You are logged in as <span class="upper">%v</span> with access level <span class="upper">%v</span></p><hr>`, usr, prefs.Accesslevels[alevel.(int)])

	servername, _ := os.Hostname()

	fmt.Fprintf(w, `<p>I am a server application running on a computer called %v</p>`, servername)
	exPath := filepath.Dir(ex)
	fmt.Fprintf(w, "<p>I'm installed in the folder <strong>%v</strong></p>", exPath)

	type tabledtl struct {
		Table  string
		Csvok  bool
		Jsonok bool
	}
	tables := []tabledtl{{"boxes", true, true}, {"contents", true, true}, {"locations", true, true}, {"users", false, false}} // history omitted

	fmt.Fprint(w, "<ul>")
	for _, tab := range tables {
		sqlx := "SELECT Count(*) As Rex FROM " + tab.Table
		rex, _ := strconv.Atoi(getValueFromDB(sqlx, "Rex", "0"))
		tabname := prefs.Table_Labels[tab.Table]
		fmt.Fprintf(w, `<li>Table <span class="keydata">%v</span> has <span class="keydata">%v</span> records `, tabname, commas(rex))

		if tab.Csvok {
			fmt.Fprintf(w, ` &nbsp;&nbsp;[<a href="/csvexp?%v=%v">Save as CSV</a>]`, Param_Labels["table"], tab.Table)
		}
		if tab.Jsonok {
			fmt.Fprintf(w, ` &nbsp;&nbsp;[<a href="/jsonexp?%v=%v">Save as JSON</a>]`, Param_Labels["table"], tab.Table)
		}

		fmt.Fprint(w, `</li>`)
	}
	fmt.Fprint(w, "</ul><hr>")

	fmt.Fprint(w, `<p><a class="btn" href="/check">Check database</a></p>`)

}

func show_terminology(w http.ResponseWriter, r *http.Request) {
	const terms = `
	<dl class="termstable">
	<dt>Location</dt>
	<dd>A storage location, a place to store <em>boxes</em>. A warehouse, cellar, cupboard, etc.</dd>
	<dt>Owner</dt>
	<dd>The individual, entity or department responsible for or having ownership of one or more <em>files</em>. Owners are identified usually by short codes representing, say, a partner's initials or a department such as 'PAYROLL'.</dd>
	<dt>File (Contents)</dt>
	<dd>A folder of related documents, belonging to a particular <em>owner</em> and <em>client</em>.</dd>
	<dt>Box</dt>
	<dd>A box or container holding one or more <em>files</em>, stored in a <em>location</em> often having a <em>storage reference</em> associated with that particular location. Each box is identified by a unique 'boxid'.</dd>
	<dt>Client</dt>
	<dd>Identifies which of the firm's clients or other external entities a <em>file</em> relates to. Both client numbers and names are held and are searchable.</dd>
	<dt>Review date</dt>
	<dd>The date (month &amp; year) when individual <em>files</em> should be considered for destruction or other disposal.</dd>
	<dt>Storage reference</dt>
	<dd>A unique reference assigned by the manager of a particular <em>location</em> to individual <em>boxes</em>. In a large facility this might be used to identify, say, a particular rack within a warehouse.</dd>
	</dl>
	`

	fmt.Fprint(w, terms)
}
