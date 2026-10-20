# AlumniSetu — Alumni–Student Interaction Platform

A desktop application for the **Technical Education Department, Govt. of Rajasthan**
that connects alumni with current students for mentorship, career guidance and
placement support.

> *Setu* (सेतु) means "bridge" — the platform is a bridge between alumni and students.

---

## The problem (SIH 2024, Govt. of Rajasthan)

| Problem stated | How AlumniSetu solves it |
| --- | --- |
| No centralized system for alumni data | A single SQLite database of alumni: batch, branch, company, role, industry, skills |
| No structured interaction | Mentorship requests, discussion forum, events & webinars |
| Students lack guidance | Mentor recommendations, job board, career-guidance sessions |
| Fake profiles / fraud | Role-based accounts + an admin `verified` flag on every alumni profile |
| Harmful content in forums | Every post and comment passes through a moderation filter before it is stored |

---

## Architecture (intentionally simple)

```
+-------------------------+          HTTP + JSON          +----------------------+
|   Electron desktop app  |  ------------------------->   |     Go API server    |
|   (plain HTML/CSS/JS)   |  <-------------------------   |   (net/http stdlib)  |
+-------------------------+        localhost:8080         +----------+-----------+
                                                                     |
                                                                     | database/sql
                                                                     v
                                                            +--------------------+
                                                            |  SQLite (file DB)  |
                                                            +--------------------+
```

**No frameworks on purpose.** No React, no Gin/Echo, no ORM, no bundler.
Everything is readable top-to-bottom so it can be explained in a demo.

| Layer | Technology | Why |
| --- | --- | --- |
| Desktop shell | Electron | Ships a real cross-platform desktop app |
| UI | Vanilla HTML/CSS/JS | Zero build step — edit a file, reload the window |
| API | Go `net/http` | Standard library only, one handler per endpoint |
| DB | SQLite via `modernc.org/sqlite` | Pure Go, no cgo, no server to install |
| Passwords | `bcrypt` | Two function calls, industry standard |
| Auth | Random token in a `sessions` table | Easier to explain than JWT, same effect |

---

## Project layout

```
ad_desktop_app/
├── server/                 # Go API
│   ├── main.go             # routes + server startup
│   ├── db.go               # database connection + schema
│   ├── models.go           # structs that mirror the tables
│   ├── auth.go             # register / login / me + auth middleware
│   ├── profiles.go         # profile read & update
│   ├── directory.go        # alumni directory: search + filters
│   ├── recommend.go        # mentor matching score
│   ├── mentorship.go       # mentorship requests
│   ├── events.go           # events + RSVP
│   ├── forum.go            # posts + comments
│   ├── jobs.go             # placement / job board
│   ├── moderation.go       # harmful-content filter
│   ├── helpers.go          # small JSON + request helpers
│   └── seed.go             # demo data
└── desktop/                # Electron app
    ├── main.js             # creates the window, starts the server
    ├── preload.js          # exposes the API base URL safely
    └── renderer/           # the actual UI
```

---

## Running it

### 1. Start the API server

```bash
cd server
go run .
```

Server listens on `http://localhost:8080`. On first run it creates `alumnisetu.db`
and seeds demo accounts.

### 2. Start the desktop app

```bash
cd desktop
npm install
npm start
```

### Demo logins

| Role | Email | Password |
| --- | --- | --- |
| Student | `riya@student.ac.in` | `password123` |
| Alumni | `arjun@alumni.ac.in` | `password123` |
| Admin | `admin@rtu.ac.in` | `password123` |

---

## Status

This is an in-progress build. See [`docs/ROADMAP.md`](docs/ROADMAP.md) for what is
done and what is next.
