package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func openDB(path string) {
	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		log.Fatal("cannot open database: ", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatal("cannot enable foreign keys: ", err)
	}

	createTables()
	log.Println("database ready:", path)
}

func createTables() {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		name          TEXT    NOT NULL,
		email         TEXT    NOT NULL UNIQUE,
		password_hash TEXT    NOT NULL,
		role          TEXT    NOT NULL CHECK (role IN ('student', 'alumni', 'admin')),
		verified      INTEGER NOT NULL DEFAULT 0,
		created_at    TEXT    NOT NULL
	);

	CREATE TABLE IF NOT EXISTS profiles (
		user_id         INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
		college         TEXT NOT NULL DEFAULT '',
		branch          TEXT NOT NULL DEFAULT '',
		graduation_year INTEGER NOT NULL DEFAULT 0,
		company         TEXT NOT NULL DEFAULT '',
		job_title       TEXT NOT NULL DEFAULT '',
		industry        TEXT NOT NULL DEFAULT '',
		location        TEXT NOT NULL DEFAULT '',
		skills          TEXT NOT NULL DEFAULT '',
		interests       TEXT NOT NULL DEFAULT '',
		bio             TEXT NOT NULL DEFAULT '',
		linkedin        TEXT NOT NULL DEFAULT '',
		open_to_mentor  INTEGER NOT NULL DEFAULT 0,
		updated_at      TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token      TEXT PRIMARY KEY,
		user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS events (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		title       TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		kind        TEXT NOT NULL DEFAULT 'webinar',
		starts_at   TEXT NOT NULL,
		location    TEXT NOT NULL DEFAULT '',
		link        TEXT NOT NULL DEFAULT '',
		created_by  INTEGER NOT NULL REFERENCES users(id),
		created_at  TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS event_rsvps (
		event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
		user_id  INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		PRIMARY KEY (event_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS posts (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		author_id  INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		title      TEXT NOT NULL,
		body       TEXT NOT NULL,
		category   TEXT NOT NULL DEFAULT 'general',
		flagged    INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS comments (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id    INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
		author_id  INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		body       TEXT NOT NULL,
		flagged    INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS mentorships (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		student_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		alumni_id  INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		message    TEXT NOT NULL DEFAULT '',
		status     TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
		created_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS ledger (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		action       TEXT NOT NULL,
		subject_id   INTEGER NOT NULL,
		subject_name TEXT NOT NULL,
		actor_id     INTEGER NOT NULL,
		actor_name   TEXT NOT NULL,
		details      TEXT NOT NULL DEFAULT '',
		created_at   TEXT NOT NULL,
		prev_hash    TEXT NOT NULL,
		hash         TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS notifications (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		kind       TEXT NOT NULL,
		title      TEXT NOT NULL,
		body       TEXT NOT NULL DEFAULT '',
		page       TEXT NOT NULL DEFAULT '',
		seen       INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS jobs (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		posted_by   INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		title       TEXT NOT NULL,
		company     TEXT NOT NULL,
		location    TEXT NOT NULL DEFAULT '',
		kind        TEXT NOT NULL DEFAULT 'full-time',
		description TEXT NOT NULL DEFAULT '',
		apply_link  TEXT NOT NULL DEFAULT '',
		created_at  TEXT NOT NULL
	);
	`

	if _, err := db.Exec(schema); err != nil {
		log.Fatal("cannot create tables: ", err)
	}
}
