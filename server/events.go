package main

import (
	"net/http"
	"strconv"
	"strings"
)

var eventKinds = []string{"meetup", "webinar", "panel"}

type eventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
	StartsAt    string `json:"starts_at"`
	Location    string `json:"location"`
	Link        string `json:"link"`
}

func handleCreateEvent(w http.ResponseWriter, r *http.Request, user *User) {
	var body eventRequest
	if !readJSON(w, r, &body) {
		return
	}

	body.Title = strings.TrimSpace(body.Title)
	body.StartsAt = strings.TrimSpace(body.StartsAt)

	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if body.StartsAt == "" {
		writeError(w, http.StatusBadRequest, "start date and time are required")
		return
	}
	if body.Kind == "" {
		body.Kind = "webinar"
	}
	if !contains(eventKinds, body.Kind) {
		writeError(w, http.StatusBadRequest, "kind must be meetup, webinar or panel")
		return
	}

	result, err := db.Exec(`INSERT INTO events
		(title, description, kind, starts_at, location, link, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		body.Title, body.Description, body.Kind, body.StartsAt,
		body.Location, body.Link, user.ID, now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create event")
		return
	}

	id, _ := result.LastInsertId()
	event, err := loadEvent(id, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load event")
		return
	}
	writeJSON(w, http.StatusCreated, event)
}

func handleListEvents(w http.ResponseWriter, r *http.Request, user *User) {
	where := "1 = 1"
	if r.URL.Query().Get("upcoming") == "true" {
		where = "e.starts_at >= ?"
	}

	args := []any{user.ID}
	if where != "1 = 1" {
		args = append(args, now())
	}

	rows, err := db.Query(eventSelect+` WHERE `+where+` ORDER BY e.starts_at ASC`, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load events")
		return
	}
	defer rows.Close()

	list := []Event{}
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not read events")
			return
		}
		list = append(list, event)
	}

	writeJSON(w, http.StatusOK, map[string]any{"count": len(list), "results": list})
}

func handleRSVP(w http.ResponseWriter, r *http.Request, user *User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	if _, err := loadEvent(id, user.ID); err != nil {
		writeError(w, http.StatusNotFound, "no such event")
		return
	}

	if r.Method == http.MethodDelete {
		db.Exec(`DELETE FROM event_rsvps WHERE event_id = ? AND user_id = ?`, id, user.ID)
	} else {
		db.Exec(`INSERT OR IGNORE INTO event_rsvps (event_id, user_id) VALUES (?, ?)`, id, user.ID)
	}

	event, _ := loadEvent(id, user.ID)
	writeJSON(w, http.StatusOK, event)
}

const eventSelect = `
	SELECT e.id, e.title, e.description, e.kind, e.starts_at, e.location, e.link,
	       e.created_by, creator.name, e.created_at,
	       (SELECT COUNT(*) FROM event_rsvps WHERE event_id = e.id),
	       (SELECT COUNT(*) FROM event_rsvps WHERE event_id = e.id AND user_id = ?)
	FROM events e
	JOIN users creator ON creator.id = e.created_by`

func scanEvent(row rowScanner) (Event, error) {
	var e Event
	var attending int
	err := row.Scan(&e.ID, &e.Title, &e.Description, &e.Kind, &e.StartsAt,
		&e.Location, &e.Link, &e.CreatedBy, &e.CreatorName, &e.CreatedAt,
		&e.RSVPCount, &attending)
	e.IsAttending = attending > 0
	return e, err
}

func loadEvent(eventID, viewerID int64) (Event, error) {
	return scanEvent(db.QueryRow(eventSelect+` WHERE e.id = ?`, viewerID, eventID))
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
