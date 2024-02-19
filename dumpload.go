package main

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
)

var tables = []string{"boxes", "contents", "history", "locations", "owners", "users"}

var tabcols = map[string]string{
	"boxes":     "storeref,boxid,location,overview,numdocs,min_review_date,max_review_date",
	"contents":  "id,boxid,review_date,contents,owner,name,client",
	"history":   "recordedat,userid,thesql,theresult",
	"locations": "location",
	"owners":    "owner,name",
	"users":     "userid,userpass,accesslevel",
}

type dbjson struct {
	Boxes     []table_boxes     `json:"boxes"`
	Contents  []table_contents  `json:"contents"`
	History   []table_history   `json:"history"`
	Locations []table_locations `json:"locations"`
	Owners    []table_owners    `json:"owners"`
	Users     []table_users     `json:"users"`
}

var jsondata dbjson

func loaddatabase(fromfile string) {

	f, err := os.Open(fromfile)
	checkerr(err)
	defer f.Close()
	bytes, err := io.ReadAll(f)
	checkerr(err)
	err = json.Unmarshal(bytes, &jsondata)
	checkerr(err)

	ownersLoaded := false

	DBExec("BEGIN TRANSACTION")
	for i := range jsondata.Boxes {
		storeBoxRecord(jsondata.Boxes[i])
	}
	for i := range jsondata.Contents {
		storeContentsRecord(jsondata.Contents[i])
	}
	for i := range jsondata.History {
		storeHistoryRecord(jsondata.History[i])
	}
	for i := range jsondata.Locations {
		storeLocationRecord(jsondata.Locations[i])
	}
	for i := range jsondata.Owners {
		storeOwnerRecord(jsondata.Owners[i])
		ownersLoaded = true
	}
	for i := range jsondata.Users {
		storeUserRecord(jsondata.Users[i])
	}
	if !ownersLoaded {
		buildOwnersTable()
	}
	DBExec("COMMIT")
}

func storeBoxRecord(data table_boxes) {

	sqlx := "INSERT OR REPLACE INTO boxes(storeref,boxid,location,overview,numdocs,min_review_date,max_review_date)"
	sqlx += "VALUES("
	sqlx += "'" + safesql(data.Storeref) + "'"
	sqlx += ",'" + safesql(data.Boxid) + "'"
	sqlx += ",'" + safesql(data.Location) + "'"
	sqlx += ",'" + safesql(data.Overview) + "'"
	sqlx += "," + strconv.Itoa(data.NumDocs)
	sqlx += ",'" + safesql(data.Min_Review_date) + "'"
	sqlx += ",'" + safesql(data.Max_Review_date) + "'"
	sqlx += ")"
	_, err := DBH.Exec(sqlx)
	checkerr(err)
}

func storeContentsRecord(rec table_contents) {

	sqlx := "INSERT OR REPLACE INTO contents(id,boxid,review_date,contents,owner,name,client)"
	sqlx += "VALUES("
	sqlx += strconv.Itoa(rec.Id)
	sqlx += ",'" + safesql(rec.Boxid) + "'"
	sqlx += ",'" + safesql(rec.Review_date) + "'"
	sqlx += ",'" + safesql(rec.Contents) + "'"
	sqlx += ",'" + safesql(rec.Owner) + "'"
	sqlx += ",'" + safesql(rec.Name) + "'"
	sqlx += ",'" + safesql(rec.Client) + "'"
	sqlx += ")"
	_, err := DBH.Exec(sqlx)
	checkerr(err)
}

func storeLocationRecord(data table_locations) {

	sqlx := "INSERT OR REPLACE INTO locations(location)"
	sqlx += "VALUES('" + safesql(data.Location) + "')"
	_, err := DBH.Exec(sqlx)
	checkerr(err)

}

func storeHistoryRecord(rec table_history) {

	sqlx := "INSERT OR REPLACE INTO history(recordedat,userid,thesql,theresult)"
	sqlx += "VALUES("
	sqlx += "'" + safesql(rec.Recordedat) + "'"
	sqlx += ",'" + safesql(rec.Userid) + "'"
	sqlx += ",'" + safesql(rec.TheSQL) + "'"
	sqlx += "," + strconv.Itoa(rec.TheResult)
	sqlx += ")"
	_, err := DBH.Exec(sqlx)
	checkerr(err)
}

func storeOwnerRecord(rec table_owners) {

	sqlx := "INSERT OR REPLACE INTO owners(owner,name)"
	sqlx += "VALUES("
	sqlx += "'" + safesql(rec.Owner) + "'"
	sqlx += ",'" + safesql(rec.Name) + "'"
	sqlx += ")"
	_, err := DBH.Exec(sqlx)
	checkerr(err)

}

func storeUserRecord(rec table_users) {

	sqlx := "INSERT OR REPLACE INTO users(userid,userpass,accesslevel)"
	sqlx += "VALUES("
	sqlx += "'" + safesql(rec.Userid) + "'"
	sqlx += ",'" + safesql(rec.Userpass) + "'"
	sqlx += "," + strconv.Itoa(rec.Accesslevel)
	sqlx += ")"
	_, err := DBH.Exec(sqlx)
	checkerr(err)

}

func dumpdatabase(tofile string) {

	f, err := os.Create(tofile)
	checkerr(err)
	defer f.Close()
	f.WriteString("{")

	commaNeeded := false
	for _, tab := range tables {
		x := getValueFromDB("SELECT name FROM sqlite_master WHERE name='"+tab+"'", "")
		if x == "" {
			continue
		}

		if commaNeeded {
			f.WriteString(",")
		}
		f.WriteString(`"` + tab + `"` + ":[")
		switch tab {
		case "boxes":
			putdata_boxes(f)
		case "contents":
			putdata_contents(f)
		case "history":
			putdata_history(f)
		case "locations":
			putdata_locations(f)
		case "owners":
			putdata_owners(f)
		case "users":
			putdata_users(f)
		}

		f.WriteString("]")
		commaNeeded = true

	}
	f.WriteString("}")

}

func putdata_boxes(f *os.File) {

	var bx table_boxes

	sqlx := "SELECT " + tabcols["boxes"] + " FROM boxes"
	r, err := DBH.Query(sqlx)
	checkerr(err)
	defer r.Close()
	commaNeeded := false
	for r.Next() {
		r.Scan(&bx.Storeref, &bx.Boxid, &bx.Location, &bx.Overview, &bx.NumDocs, &bx.Min_Review_date, &bx.Max_Review_date)
		b, err := json.Marshal(bx)
		checkerr(err)

		if commaNeeded {
			f.WriteString(",")
		}
		f.Write(b)
		commaNeeded = true

	}

}

func putdata_contents(f *os.File) {

	var cn table_contents

	sqlx := "SELECT " + tabcols["contents"] + " FROM contents"
	r, err := DBH.Query(sqlx)
	checkerr(err)
	defer r.Close()
	commaNeeded := false
	for r.Next() {
		r.Scan(&cn.Id, &cn.Boxid, &cn.Review_date, &cn.Contents, &cn.Owner, &cn.Name, &cn.Client)
		b, err := json.Marshal(cn)
		checkerr(err)

		if commaNeeded {
			f.WriteString(",")
		}
		f.Write(b)
		commaNeeded = true

	}

}

func putdata_history(f *os.File) {

	var hr table_history

	sqlx := "SELECT " + tabcols["history"] + " FROM history"
	r, err := DBH.Query(sqlx)
	checkerr(err)
	defer r.Close()
	commaNeeded := false
	for r.Next() {
		r.Scan(&hr.Recordedat, &hr.Userid, &hr.TheSQL, &hr.TheResult)
		b, err := json.Marshal(hr)
		checkerr(err)

		if commaNeeded {
			f.WriteString(",")
		}
		f.Write(b)
		commaNeeded = true

	}

}

func putdata_locations(f *os.File) {

	var loc table_locations

	sqlx := "SELECT " + tabcols["locations"] + " FROM locations"
	r, err := DBH.Query(sqlx)
	checkerr(err)
	defer r.Close()
	commaNeeded := false
	for r.Next() {
		r.Scan(&loc.Location)
		b, err := json.Marshal(loc)
		checkerr(err)

		if commaNeeded {
			f.WriteString(",")
		}
		f.Write(b)
		commaNeeded = true

	}

}

func putdata_owners(f *os.File) {

	var own table_owners

	sqlx := "SELECT " + tabcols["owners"] + " FROM owners"
	r, err := DBH.Query(sqlx)
	checkerr(err)
	defer r.Close()
	commaNeeded := false
	for r.Next() {
		r.Scan(&own.Owner, &own.Name)
		b, err := json.Marshal(own)
		checkerr(err)

		if commaNeeded {
			f.WriteString(",")
		}
		f.Write(b)
		commaNeeded = true

	}

}

func putdata_users(f *os.File) {

	var usr table_users

	sqlx := "SELECT " + tabcols["users"] + " FROM users"
	r, err := DBH.Query(sqlx)
	checkerr(err)
	defer r.Close()
	commaNeeded := false
	for r.Next() {
		r.Scan(&usr.Userid, &usr.Userpass, &usr.Accesslevel)
		b, err := json.Marshal(usr)
		checkerr(err)

		if commaNeeded {
			f.WriteString(",")
		}
		f.Write(b)
		commaNeeded = true

	}

}
