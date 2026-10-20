package main

import (
	"net/http"
	"strconv"
	"strings"
)

type mentorshipRequest struct {
	AlumniID int64  `json:"alumni_id"`
	Message  string `json:"message"`
}

func handleCreateMentorship(w http.ResponseWriter, r *http.Request, user *User) {
	var body mentorshipRequest
	if !readJSON(w, r, &body) {
		return
	}

	alumni, err := findUserByID(body.AlumniID)
	if err != nil || alumni.Role != "alumni" {
		writeError(w, http.StatusBadRequest, "that mentor does not exist")
		return
	}

	var existing int
	db.QueryRow(`SELECT COUNT(*) FROM mentorships
		WHERE student_id = ? AND alumni_id = ? AND status = 'pending'`,
		user.ID, alumni.ID).Scan(&existing)
	if existing > 0 {
		writeError(w, http.StatusConflict, "you already have a pending request with this mentor")
		return
	}

	result, err := db.Exec(`INSERT INTO mentorships (student_id, alumni_id, message, status, created_at)
		VALUES (?, ?, ?, 'pending', ?)`,
		user.ID, alumni.ID, strings.TrimSpace(body.Message), now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not send request")
		return
	}

	id, _ := result.LastInsertId()
	mentorship, err := loadMentorship(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load request")
		return
	}
	writeJSON(w, http.StatusCreated, mentorship)
}

func handleListMentorships(w http.ResponseWriter, r *http.Request, user *User) {
	rows, err := db.Query(mentorshipSelect+`
		WHERE m.student_id = ? OR m.alumni_id = ?
		ORDER BY m.created_at DESC`, user.ID, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load requests")
		return
	}
	defer rows.Close()

	list := []Mentorship{}
	for rows.Next() {
		m, err := scanMentorship(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not read requests")
			return
		}
		list = append(list, m)
	}

	writeJSON(w, http.StatusOK, map[string]any{"count": len(list), "results": list})
}

type mentorshipStatusRequest struct {
	Status string `json:"status"`
}

func handleUpdateMentorship(w http.ResponseWriter, r *http.Request, user *User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request id")
		return
	}

	var body mentorshipStatusRequest
	if !readJSON(w, r, &body) {
		return
	}
	if body.Status != "accepted" && body.Status != "rejected" {
		writeError(w, http.StatusBadRequest, "status must be 'accepted' or 'rejected'")
		return
	}

	mentorship, err := loadMentorship(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "no such request")
		return
	}
	if mentorship.AlumniID != user.ID {
		writeError(w, http.StatusForbidden, "only the mentor can answer this request")
		return
	}
	if mentorship.Status != "pending" {
		writeError(w, http.StatusConflict, "this request was already answered")
		return
	}

	if _, err := db.Exec(`UPDATE mentorships SET status = ? WHERE id = ?`, body.Status, id); err != nil {
		writeError(w, http.StatusInternalServerError, "could not update request")
		return
	}

	updated, _ := loadMentorship(id)
	writeJSON(w, http.StatusOK, updated)
}

const mentorshipSelect = `
	SELECT m.id, m.student_id, student.name, m.alumni_id, alumni.name,
	       m.message, m.status, m.created_at
	FROM mentorships m
	JOIN users student ON student.id = m.student_id
	JOIN users alumni  ON alumni.id  = m.alumni_id`

func scanMentorship(row rowScanner) (Mentorship, error) {
	var m Mentorship
	err := row.Scan(&m.ID, &m.StudentID, &m.StudentName, &m.AlumniID,
		&m.AlumniName, &m.Message, &m.Status, &m.CreatedAt)
	return m, err
}

func loadMentorship(id int64) (Mentorship, error) {
	return scanMentorship(db.QueryRow(mentorshipSelect+` WHERE m.id = ?`, id))
}
