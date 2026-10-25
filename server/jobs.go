package main

import (
	"net/http"
	"strconv"
	"strings"
)

var jobKinds = []string{"full-time", "internship"}

type jobRequest struct {
	Title       string `json:"title"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
	ApplyLink   string `json:"apply_link"`
}

func handleCreateJob(w http.ResponseWriter, r *http.Request, user *User) {
	var body jobRequest
	if !readJSON(w, r, &body) {
		return
	}

	body.Title = strings.TrimSpace(body.Title)
	body.Company = strings.TrimSpace(body.Company)

	if body.Title == "" || body.Company == "" {
		writeError(w, http.StatusBadRequest, "title and company are required")
		return
	}
	if body.Kind == "" {
		body.Kind = "full-time"
	}
	if !contains(jobKinds, body.Kind) {
		writeError(w, http.StatusBadRequest, "kind must be full-time or internship")
		return
	}

	body.Description, _ = moderate(body.Description)

	result, err := db.Exec(`INSERT INTO jobs
		(posted_by, title, company, location, kind, description, apply_link, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID, body.Title, body.Company, body.Location, body.Kind,
		body.Description, body.ApplyLink, now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not post job")
		return
	}

	id, _ := result.LastInsertId()
	job, err := loadJob(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load job")
		return
	}

	notifyEveryoneExcept(user.ID, "job",
		"New opening: "+job.Title+" at "+job.Company,
		job.Location, "jobs")

	writeJSON(w, http.StatusCreated, job)
}

func handleListJobs(w http.ResponseWriter, r *http.Request, _ *User) {
	where := []string{"1 = 1"}
	args := []any{}

	if kind := r.URL.Query().Get("kind"); kind != "" {
		where = append(where, "j.kind = ?")
		args = append(args, kind)
	}
	if q := strings.TrimSpace(r.URL.Query().Get("q")); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		where = append(where, `(LOWER(j.title) LIKE ? OR LOWER(j.company) LIKE ?
			OR LOWER(j.location) LIKE ? OR LOWER(j.description) LIKE ?)`)
		args = append(args, like, like, like, like)
	}

	rows, err := db.Query(jobSelect+` WHERE `+strings.Join(where, " AND ")+
		` ORDER BY j.created_at DESC LIMIT 100`, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load jobs")
		return
	}
	defer rows.Close()

	list := []Job{}
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not read jobs")
			return
		}
		list = append(list, job)
	}

	writeJSON(w, http.StatusOK, map[string]any{"count": len(list), "results": list})
}

func handleDeleteJob(w http.ResponseWriter, r *http.Request, user *User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	job, err := loadJob(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "no such job")
		return
	}
	if job.PostedBy != user.ID && user.Role != "admin" {
		writeError(w, http.StatusForbidden, "you can only remove jobs you posted")
		return
	}

	db.Exec(`DELETE FROM jobs WHERE id = ?`, id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

const jobSelect = `
	SELECT j.id, j.posted_by, poster.name, j.title, j.company, j.location,
	       j.kind, j.description, j.apply_link, j.created_at
	FROM jobs j
	JOIN users poster ON poster.id = j.posted_by`

func scanJob(row rowScanner) (Job, error) {
	var j Job
	err := row.Scan(&j.ID, &j.PostedBy, &j.PosterName, &j.Title, &j.Company,
		&j.Location, &j.Kind, &j.Description, &j.ApplyLink, &j.CreatedAt)
	return j, err
}

func loadJob(id int64) (Job, error) {
	return scanJob(db.QueryRow(jobSelect+` WHERE j.id = ?`, id))
}
