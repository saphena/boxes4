-- This script will upgrade boxes.db running live in SS to Boxes4 standard
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "owners" (
	"owner" TEXT NOT NULL COLLATE NOCASE,
	"name" TEXT COLLATE NOCASE,
	PRIMARY KEY("owner")
);
INSERT OR IGNORE INTO users (userid,userpass,accesslevel) VALUES('admin','admin',9);
COMMIT;