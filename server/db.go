package main

// db.go — opens the SQLite file and creates every table the app needs.
//
// We use modernc.org/sqlite, which is SQLite rewritten in pure Go. That means
// no C compiler and no database server to install: the whole database is one
// file called alumnisetu.db sitting next to the program.

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite" // registers the "sqlite" driver with database/sql
)

// db is the one shared database handle. database/sql already handles
// connection pooling and is safe to use from many requests at once, so a
// single global is all we need.
var db *sql.DB

// openDB connects to the database file and makes sure all tables exist.
func openDB(path string) {
	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		log.Fatal("cannot open database: ", err)
	}

	// Foreign keys are off by default in SQLite. Turn them on so the database
	// itself refuses rows that point at users or posts that do not exist.
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatal("cannot enable foreign keys: ", err)
	}

	createTables()
	log.Println("database ready:", path)
}

// createTables runs the schema. Every statement uses IF NOT EXISTS, so calling
// this on every startup is harmless.
func createTables() {
	schema := `
	-- Everyone who can log in: students, alumni and department admins.
	CREATE TABLE IF NOT EXISTS users (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		name          TEXT    NOT NULL,
		email         TEXT    NOT NULL UNIQUE,
		password_hash TEXT    NOT NULL,
		role          TEXT    NOT NULL CHECK (role IN ('student', 'alumni', 'admin')),
		-- verified is the anti-fake-profile flag. An admin flips it to 1 once
		-- the person's degree / employment has been checked.
		verified      INTEGER NOT NULL DEFAULT 0,
		created_at    TEXT    NOT NULL
	);

	-- The extra details that make the alumni database useful. One row per user.
	CREATE TABLE IF NOT EXISTS profiles (
		user_id         INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
		college         TEXT NOT NULL DEFAULT '',
		branch          TEXT NOT NULL DEFAULT '',
		graduation_year INTEGER NOT NULL DEFAULT 0,
		company         TEXT NOT NULL DEFAULT '',
		job_title       TEXT NOT NULL DEFAULT '',
		industry        TEXT NOT NULL DEFAULT '',
		location        TEXT NOT NULL DEFAULT '',
		-- skills and interests are stored as simple comma separated text,
		-- e.g. "Go, React, Machine Learning". Easy to read, easy to search.
		skills          TEXT NOT NULL DEFAULT '',
		interests       TEXT NOT NULL DEFAULT '',
		bio             TEXT NOT NULL DEFAULT '',
		linkedin        TEXT NOT NULL DEFAULT '',
		-- Alumni set this to 1 to appear in mentor recommendations.
		open_to_mentor  INTEGER NOT NULL DEFAULT 0,
		updated_at      TEXT NOT NULL DEFAULT ''
	);

	-- A login token. Logging in inserts a row, logging out deletes it.
	CREATE TABLE IF NOT EXISTS sessions (
		token      TEXT PRIMARY KEY,
		user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TEXT NOT NULL
	);

	-- Alumni meetups, webinars and panel discussions.
	CREATE TABLE IF NOT EXISTS events (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		title       TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		kind        TEXT NOT NULL DEFAULT 'webinar', -- meetup | webinar | panel
		starts_at   TEXT NOT NULL,
		location    TEXT NOT NULL DEFAULT '',
		link        TEXT NOT NULL DEFAULT '',
		created_by  INTEGER NOT NULL REFERENCES users(id),
		created_at  TEXT NOT NULL
	);

	-- Who is attending which event. The composite key stops double booking.
	CREATE TABLE IF NOT EXISTS event_rsvps (
		event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
		user_id  INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		PRIMARY KEY (event_id, user_id)
	);

	-- Discussion forum threads.
	CREATE TABLE IF NOT EXISTS posts (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		author_id  INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		title      TEXT NOT NULL,
		body       TEXT NOT NULL,
		category   TEXT NOT NULL DEFAULT 'general', -- general | career | academic | placement
		flagged    INTEGER NOT NULL DEFAULT 0,      -- set by the moderation filter
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

	-- A student asking an alumnus to mentor them.
	CREATE TABLE IF NOT EXISTS mentorships (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		student_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		alumni_id  INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		message    TEXT NOT NULL DEFAULT '',
		status     TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
		created_at TEXT NOT NULL
	);

	-- Placement assistance: jobs and internships posted by alumni.
	CREATE TABLE IF NOT EXISTS jobs (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		posted_by   INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		title       TEXT NOT NULL,
		company     TEXT NOT NULL,
		location    TEXT NOT NULL DEFAULT '',
		kind        TEXT NOT NULL DEFAULT 'full-time', -- full-time | internship
		description TEXT NOT NULL DEFAULT '',
		apply_link  TEXT NOT NULL DEFAULT '',
		created_at  TEXT NOT NULL
	);
	`

	if _, err := db.Exec(schema); err != nil {
		log.Fatal("cannot create tables: ", err)
	}
}
