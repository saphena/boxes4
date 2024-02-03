-- This script will upgrade boxes.db running live in SS to Boxes4 standard
BEGIN TRANSACTION;
INSERT OR IGNORE INTO users (userid,userpass,accesslevel) VALUES('admin','admin',9);
COMMIT;