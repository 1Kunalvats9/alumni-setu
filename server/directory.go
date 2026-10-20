package main

// directory.go — the searchable alumni directory.
//
// GET /api/directory?q=google&role=alumni&branch=Computer%20Science&year=2016
//
// The handler builds the WHERE clause one condition at a time. Each condition
// adds a "?" placeholder and its value to a slice, so user input is always
// passed as a bound parameter and never pasted into the SQL string. That is
// what makes this safe from SQL injection.

import (
	"net/http"
	"strconv"
	"strings"
)

func handleDirectory(w http.ResponseWriter, r *http.Request, _ *User) {
	query := r.URL.Query()

	// Start with conditions that are always true, then narrow down.
	where := []string{"1 = 1"}
	args := []any{}

	// Free text search across the fields a student would actually type into
	// a search box: name, company, job title, skills, branch.
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

	// Exact-match dropdown filters.
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

	// ?mentor=true shows only alumni who ticked "open to mentoring".
	if query.Get("mentor") == "true" {
		where = append(where, "p.open_to_mentor = 1")
	}
	// ?verified=true shows only admin-verified accounts.
	if query.Get("verified") == "true" {
		where = append(where, "u.verified = 1")
	}

	// Cap the page size so one request can never pull the whole database.
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

	// Start as an empty slice, not nil, so an empty result encodes as [] and
	// not null — the UI can then loop over it without a special case.
	entries := []DirectoryEntry{}
	for rows.Next() {
		var e DirectoryEntry
		var p Profile
		err := rows.Scan(
			&e.ID, &e.Name, &e.Email, &e.Role, &e.Verified, &e.CreatedAt,
			&p.UserID, &p.College, &p.Branch, &p.GraduationYear, &p.Company,
			&p.JobTitle, &p.Industry, &p.Location, &p.Skills, &p.Interests,
			&p.Bio, &p.LinkedIn, &p.OpenToMentor, &p.UpdatedAt,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not read results")
			return
		}
		e.Profile = p
		entries = append(entries, e)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"count":   len(entries),
		"results": entries,
	})
}

// handleFilterOptions feeds the dropdowns in the UI: every branch, industry
// and batch year that actually exists in the database, so the filters can
// never offer an option that returns nothing.
func handleFilterOptions(w http.ResponseWriter, r *http.Request, _ *User) {
	writeJSON(w, http.StatusOK, map[string]any{
		"branches":   distinctValues("branch"),
		"industries": distinctValues("industry"),
		"companies":  distinctValues("company"),
		"locations":  distinctValues("location"),
		"years":      distinctValues("graduation_year"),
	})
}

// distinctValues returns every non-empty value of one profile column.
// The column name is chosen by us in the line above, never by the user, so
// putting it straight into the SQL text is safe here.
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
