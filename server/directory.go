package main

import (
	"net/http"
	"strconv"
	"strings"
)

func handleDirectory(w http.ResponseWriter, r *http.Request, _ *User) {
	query := r.URL.Query()

	where := []string{"1 = 1"}
	args := []any{}

	if q := strings.TrimSpace(query.Get("q")); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		where = append(where, `(
			LOWER(u.name)       LIKE ? OR
			LOWER(p.company)    LIKE ? OR
			LOWER(p.job_title)  LIKE ? OR
			LOWER(p.skills)     LIKE ? OR
			LOWER(p.industry)   LIKE ? OR
			LOWER(p.branch)     LIKE ?
		)`)
		args = append(args, like, like, like, like, like, like)
	}

	addFilter := func(param, column string) {
		if value := strings.TrimSpace(query.Get(param)); value != "" {
			where = append(where, "LOWER("+column+") = ?")
			args = append(args, strings.ToLower(value))
		}
	}
	addFilter("role", "u.role")
	addFilter("branch", "p.branch")
	addFilter("industry", "p.industry")
	addFilter("company", "p.company")
	addFilter("location", "p.location")

	if year := query.Get("year"); year != "" {
		if n, err := strconv.Atoi(year); err == nil {
			where = append(where, "p.graduation_year = ?")
			args = append(args, n)
		}
	}

	if query.Get("mentor") == "true" {
		where = append(where, "p.open_to_mentor = 1")
	}
	if query.Get("verified") == "true" {
		where = append(where, "u.verified = 1")
	}

	limit := 50
	if n, err := strconv.Atoi(query.Get("limit")); err == nil && n > 0 && n <= 200 {
		limit = n
	}
	offset, _ := strconv.Atoi(query.Get("offset"))
	if offset < 0 {
		offset = 0
	}

	sqlText := `
		SELECT u.id, u.name, u.email, u.role, u.verified, u.created_at, ` +
		profileColumns + `
		FROM users u
		JOIN profiles p ON p.user_id = u.id
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY u.verified DESC, p.graduation_year DESC, u.name ASC
		LIMIT ? OFFSET ?`

	rows, err := db.Query(sqlText, append(args, limit, offset)...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "directory search failed")
		return
	}
	defer rows.Close()

	entries := []DirectoryEntry{}
	for rows.Next() {
		entry, err := scanDirectoryRow(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not read results")
			return
		}
		entries = append(entries, entry)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"count":   len(entries),
		"results": entries,
	})
}

func scanDirectoryRow(row rowScanner) (DirectoryEntry, error) {
	var e DirectoryEntry
	var p Profile
	err := row.Scan(
		&e.ID, &e.Name, &e.Email, &e.Role, &e.Verified, &e.CreatedAt,
		&p.UserID, &p.College, &p.Branch, &p.GraduationYear, &p.Company,
		&p.JobTitle, &p.Industry, &p.Location, &p.Skills, &p.Interests,
		&p.Bio, &p.LinkedIn, &p.OpenToMentor, &p.UpdatedAt,
	)
	e.Profile = p
	return e, err
}

func handleFilterOptions(w http.ResponseWriter, r *http.Request, _ *User) {
	writeJSON(w, http.StatusOK, map[string]any{
		"branches":   distinctValues("branch"),
		"industries": distinctValues("industry"),
		"companies":  distinctValues("company"),
		"locations":  distinctValues("location"),
		"years":      distinctValues("graduation_year"),
	})
}

func distinctValues(column string) []string {
	rows, err := db.Query(`SELECT DISTINCT ` + column + ` FROM profiles
		WHERE ` + column + ` != '' AND ` + column + ` != 0
		ORDER BY ` + column)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	values := []string{}
	for rows.Next() {
		var v string
		if rows.Scan(&v) == nil {
			values = append(values, v)
		}
	}
	return values
}
