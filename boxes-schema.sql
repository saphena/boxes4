BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "history" (
	"recordedat"	TIMESTAMP NOT NULL,
	"userid"	TEXT NOT NULL,
	"thesql"	TEXT NOT NULL,
	"theresult"	INTEGER NOT NULL,
	"recid"	INTEGER NOT NULL DEFAULT 1,
	PRIMARY KEY("recid")
);
CREATE TABLE IF NOT EXISTS "boxes" (
	"storeref"	TEXT COLLATE NOCASE,
	"boxid"	TEXT NOT NULL COLLATE NOCASE,
	"location"	TEXT COLLATE NOCASE,
	"overview"	TEXT COLLATE NOCASE,
	"numdocs"	INTEGER,
	"min_review_date"	TEXT COLLATE NOCASE,
	"max_review_date"	TEXT COLLATE NOCASE,
	PRIMARY KEY("boxid")
);
CREATE TABLE IF NOT EXISTS "users" (
	"userid"	TEXT NOT NULL COLLATE NOCASE,
	"userpass"	TEXT,
	"accesslevel"	INTEGER,
	PRIMARY KEY("userid")
);
CREATE TABLE IF NOT EXISTS "locations" (
	"location"	TEXT NOT NULL COLLATE NOCASE,
	PRIMARY KEY("location")
);
CREATE TABLE IF NOT EXISTS "contents" (
	"id"	INTEGER NOT NULL,
	"boxid"	TEXT NOT NULL COLLATE NOCASE,
	"review_date"	TEXT COLLATE NOCASE,
	"contents"	TEXT COLLATE NOCASE,
	"owner"	TEXT COLLATE NOCASE,
	"name"	TEXT COLLATE NOCASE,
	"client"	TEXT COLLATE NOCASE,
	PRIMARY KEY("id")
);
CREATE TABLE IF NOT EXISTS "owners" (
	"owner" TEXT NOT NULL COLLATE NOCASE,
	"name" TEXT COLLATE NOCASE,
	PRIMARY KEY("owner")
);
INSERT OR IGNORE INTO locations (location) VALUES('default');
INSERT OR IGNORE INTO users (userid,userpass,accesslevel) VALUES('admin','admin',9);
COMMIT;
