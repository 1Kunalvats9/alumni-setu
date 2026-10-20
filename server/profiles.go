package main

import (
	"database/sql"
	"net/http"
	"strconv"
)

const profileColumns = `user_id, college, branch, graduation_year, company,
	job_title, industry, location, skills, interests, bio, linkedin,
	open_to_mentor, updated_at`

type rowScanner interface{ Scan(dest ...any) error }

func scanProfile(row rowScanner) (Profile, error) {
	var p Profile
	err := row.Scan(&p.UserID, &p.College, &p.Branch, &p.GraduationYear,
		&p.Company, &p.JobTitle, &p.Industry, &p.Location, &p.Skills,
		&p.Interests, &p.Bio, &p.LinkedIn, &p.OpenToMentor, &p.UpdatedAt)
	return p, err
}

func loadProfile(userID int64) (Profile, error) {
	row := db.QueryRow(`SELECT `+profileColumns+` FROM profiles WHERE user_id = ?`, userID)
	profile, err := scanProfile(row)
	if err == sql.ErrNoRows {
		return Profile{UserID: userID}, nil
	}
	return profile, err
}

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

	if body.GraduationYear != 0 && (body.GraduationYear < 1950 || body.GraduationYear > 2100) {
		writeError(w, http.StatusBadRequest, "graduation year looks wrong")
		return
	}

	body.Skills = tidyList(body.Skills)
	body.Interests = tidyList(body.Interests)
	body.Bio, _ = moderate(body.Bio)

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
