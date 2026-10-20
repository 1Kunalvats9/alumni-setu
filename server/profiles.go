package main

// profiles.go — reading and updating the extra details that turn a login
// account into a useful alumni record: batch, branch, company, skills and so
// on. This is the data the whole platform is built on, so it lives in its own
// table and has its own endpoints.

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
)

// profileColumns is written once and reused by every SELECT below, so the
// column order always matches scanProfile.
const profileColumns = `user_id, college, branch, graduation_year, company,
	job_title, industry, location, skills, interests, bio, linkedin,
	open_to_mentor, updated_at`

// scanProfile copies one database row into a Profile struct. sql.Row and
// sql.Rows both satisfy this interface, so the same function works for a
// single lookup and for a list.
type rowScanner interface{ Scan(dest ...any) error }

func scanProfile(row rowScanner) (Profile, error) {
	var p Profile
	err := row.Scan(&p.UserID, &p.College, &p.Branch, &p.GraduationYear,
		&p.Company, &p.JobTitle, &p.Industry, &p.Location, &p.Skills,
		&p.Interests, &p.Bio, &p.LinkedIn, &p.OpenToMentor, &p.UpdatedAt)
	return p, err
}

// loadProfile fetches one user's profile. A user created before this table
// existed might not have a row yet, so a missing row returns an empty profile
// rather than an error.
func loadProfile(userID int64) (Profile, error) {
	row := db.QueryRow(`SELECT `+profileColumns+` FROM profiles WHERE user_id = ?`, userID)
	profile, err := scanProfile(row)
	if err == sql.ErrNoRows {
		return Profile{UserID: userID}, nil
	}
	return profile, err
}

// --- GET /api/users/{id} -------------------------------------------------

// handleGetUser returns one person's public card: their user row plus their
// profile. Used when a student clicks someone in the directory.
func handleGetUser(w http.ResponseWriter, r *http.Request, _ *User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := findUserByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "no such user")
		return
	}
	profile, err := loadProfile(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load profile")
		return
	}

	writeJSON(w, http.StatusOK, DirectoryEntry{User: *user, Profile: profile})
}

// --- PUT /api/me/profile -------------------------------------------------

// profileRequest is what the desktop app sends when the user saves the
// profile form. Every field is sent every time, so the handler is a single
// UPDATE with no "did they change this one?" logic.
type profileRequest struct {
	College        string `json:"college"`
	Branch         string `json:"branch"`
	GraduationYear int    `json:"graduation_year"`
	Company        string `json:"company"`
	JobTitle       string `json:"job_title"`
	Industry       string `json:"industry"`
	Location       string `json:"location"`
	Skills         string `json:"skills"`
	Interests      string `json:"interests"`
	Bio            string `json:"bio"`
	LinkedIn       string `json:"linkedin"`
	OpenToMentor   bool   `json:"open_to_mentor"`
}

func handleUpdateProfile(w http.ResponseWriter, r *http.Request, user *User) {
	var body profileRequest
	if !readJSON(w, r, &body) {
		return
	}

	// Graduation year sanity check: catches typos like 201 or 20255.
	if body.GraduationYear != 0 && (body.GraduationYear < 1950 || body.GraduationYear > 2100) {
		writeError(w, http.StatusBadRequest, "graduation year looks wrong")
		return
	}

	// Tidy the comma separated lists so "Go,  react ,ML" is stored as
	// "Go, react, ML" and searching stays predictable.
	body.Skills = tidyList(body.Skills)
	body.Interests = tidyList(body.Interests)

	// INSERT ... ON CONFLICT means this one statement creates the row if it
	// is missing and updates it if it already exists.
	_, err := db.Exec(`
		INSERT INTO profiles (user_id, college, branch, graduation_year, company,
			job_title, industry, location, skills, interests, bio, linkedin,
			open_to_mentor, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			college = excluded.college,
			branch = excluded.branch,
			graduation_year = excluded.graduation_year,
			company = excluded.company,
			job_title = excluded.job_title,
			industry = excluded.industry,
			location = excluded.location,
			skills = excluded.skills,
			interests = excluded.interests,
			bio = excluded.bio,
			linkedin = excluded.linkedin,
			open_to_mentor = excluded.open_to_mentor,
			updated_at = excluded.updated_at`,
		user.ID, body.College, body.Branch, body.GraduationYear, body.Company,
		body.JobTitle, body.Industry, body.Location, body.Skills, body.Interests,
		body.Bio, body.LinkedIn, body.OpenToMentor, now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not save profile")
		return
	}

	profile, _ := loadProfile(user.ID)
	writeJSON(w, http.StatusOK, profile)
}

// tidyList normalises "Go,  react ,,ML " into "Go, react, ML".
func tidyList(csv string) string {
	var parts []string
	for _, part := range strings.Split(csv, ",") {
		if part = strings.TrimSpace(part); part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ", ")
}
