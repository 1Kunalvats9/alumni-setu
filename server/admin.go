package main

import (
	"net/http"
	"strconv"
)

type platformStats struct {
	Students           int `json:"students"`
	Alumni             int `json:"alumni"`
	VerifiedAlumni     int `json:"verified_alumni"`
	PendingAlumni      int `json:"pending_alumni"`
	OpenToMentor       int `json:"open_to_mentor"`
	Events             int `json:"events"`
	RSVPs              int `json:"rsvps"`
	Posts              int `json:"posts"`
	Comments           int `json:"comments"`
	FlaggedPosts       int `json:"flagged_posts"`
	FlaggedComments    int `json:"flagged_comments"`
	Jobs               int `json:"jobs"`
	MentorshipsPending int `json:"mentorships_pending"`
	MentorshipsActive  int `json:"mentorships_active"`
}

func countRows(query string, args ...any) int {
	var n int
	db.QueryRow(query, args...).Scan(&n)
	return n
}

func handleAdminStats(w http.ResponseWriter, r *http.Request, _ *User) {
	stats := platformStats{
		Students:           countRows(`SELECT COUNT(*) FROM users WHERE role = 'student'`),
		Alumni:             countRows(`SELECT COUNT(*) FROM users WHERE role = 'alumni'`),
		VerifiedAlumni:     countRows(`SELECT COUNT(*) FROM users WHERE role = 'alumni' AND verified = 1`),
		PendingAlumni:      countRows(`SELECT COUNT(*) FROM users WHERE role = 'alumni' AND verified = 0`),
		OpenToMentor:       countRows(`SELECT COUNT(*) FROM profiles WHERE open_to_mentor = 1`),
		Events:             countRows(`SELECT COUNT(*) FROM events`),
		RSVPs:              countRows(`SELECT COUNT(*) FROM event_rsvps`),
		Posts:              countRows(`SELECT COUNT(*) FROM posts`),
		Comments:           countRows(`SELECT COUNT(*) FROM comments`),
		FlaggedPosts:       countRows(`SELECT COUNT(*) FROM posts WHERE flagged = 1`),
		FlaggedComments:    countRows(`SELECT COUNT(*) FROM comments WHERE flagged = 1`),
		Jobs:               countRows(`SELECT COUNT(*) FROM jobs`),
		MentorshipsPending: countRows(`SELECT COUNT(*) FROM mentorships WHERE status = 'pending'`),
		MentorshipsActive:  countRows(`SELECT COUNT(*) FROM mentorships WHERE status = 'accepted'`),
	}

	writeJSON(w, http.StatusOK, stats)
}

func handleAdminPending(w http.ResponseWriter, r *http.Request, _ *User) {
	rows, err := db.Query(`
		SELECT u.id, u.name, u.email, u.role, u.verified, u.created_at, ` + profileColumns + `
		FROM users u
		JOIN profiles p ON p.user_id = u.id
		WHERE u.role = 'alumni' AND u.verified = 0
		ORDER BY u.created_at ASC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load pending accounts")
		return
	}
	defer rows.Close()

	entries := []DirectoryEntry{}
	for rows.Next() {
		entry, err := scanDirectoryRow(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not read accounts")
			return
		}
		entries = append(entries, entry)
	}

	writeJSON(w, http.StatusOK, map[string]any{"count": len(entries), "results": entries})
}

type verifyRequest struct {
	Verified bool `json:"verified"`
}

func handleAdminVerify(w http.ResponseWriter, r *http.Request, _ *User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var body verifyRequest
	if !readJSON(w, r, &body) {
		return
	}

	target, err := findUserByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "no such user")
		return
	}
	if target.Role == "admin" {
		writeError(w, http.StatusForbidden, "admin accounts cannot be changed here")
		return
	}

	if _, err := db.Exec(`UPDATE users SET verified = ? WHERE id = ?`, body.Verified, id); err != nil {
		writeError(w, http.StatusInternalServerError, "could not update account")
		return
	}

	updated, _ := findUserByID(id)
	writeJSON(w, http.StatusOK, updated)
}

func handleAdminFlagged(w http.ResponseWriter, r *http.Request, _ *User) {
	postRows, err := db.Query(postSelect + ` WHERE p.flagged = 1 ORDER BY p.created_at DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load flagged posts")
		return
	}
	defer postRows.Close()

	posts := []Post{}
	for postRows.Next() {
		post, err := scanPost(postRows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not read flagged posts")
			return
		}
		posts = append(posts, post)
	}

	commentRows, err := db.Query(commentSelect + ` WHERE c.flagged = 1 ORDER BY c.created_at DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load flagged comments")
		return
	}
	defer commentRows.Close()

	comments := []Comment{}
	for commentRows.Next() {
		comment, err := scanComment(commentRows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not read flagged comments")
			return
		}
		comments = append(comments, comment)
	}

	writeJSON(w, http.StatusOK, map[string]any{"posts": posts, "comments": comments})
}

func handleAdminClearFlag(w http.ResponseWriter, r *http.Request, _ *User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	if _, err := loadPost(id); err != nil {
		writeError(w, http.StatusNotFound, "no such post")
		return
	}

	db.Exec(`UPDATE posts SET flagged = 0 WHERE id = ?`, id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "approved"})
}
