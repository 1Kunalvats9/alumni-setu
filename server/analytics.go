package main

import "net/http"

type Bucket struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

func groupCount(query string, args ...any) []Bucket {
	rows, err := db.Query(query, args...)
	if err != nil {
		return []Bucket{}
	}
	defer rows.Close()

	buckets := []Bucket{}
	for rows.Next() {
		var b Bucket
		if rows.Scan(&b.Label, &b.Value) == nil {
			buckets = append(buckets, b)
		}
	}
	return buckets
}

func handleAnalytics(w http.ResponseWriter, r *http.Request, _ *User) {
	writeJSON(w, http.StatusOK, map[string]any{
		"alumni_by_batch": groupCount(`
			SELECT CAST(p.graduation_year AS TEXT), COUNT(*)
			FROM users u JOIN profiles p ON p.user_id = u.id
			WHERE u.role = 'alumni' AND p.graduation_year > 0
			GROUP BY p.graduation_year ORDER BY p.graduation_year ASC`),

		"alumni_by_industry": groupCount(`
			SELECT p.industry, COUNT(*)
			FROM users u JOIN profiles p ON p.user_id = u.id
			WHERE u.role = 'alumni' AND p.industry != ''
			GROUP BY p.industry ORDER BY COUNT(*) DESC, p.industry ASC LIMIT 8`),

		"members_by_branch": groupCount(`
			SELECT p.branch, COUNT(*)
			FROM users u JOIN profiles p ON p.user_id = u.id
			WHERE p.branch != ''
			GROUP BY p.branch ORDER BY COUNT(*) DESC, p.branch ASC LIMIT 8`),

		"top_mentors": groupCount(`
			SELECT u.name, COUNT(*)
			FROM mentorships m JOIN users u ON u.id = m.alumni_id
			WHERE m.status = 'accepted'
			GROUP BY m.alumni_id ORDER BY COUNT(*) DESC, u.name ASC LIMIT 8`),

		"sessions_by_kind": groupCount(`
			SELECT kind, COUNT(*) FROM events
			GROUP BY kind ORDER BY COUNT(*) DESC, kind ASC`),

		"posts_by_category": groupCount(`
			SELECT category, COUNT(*) FROM posts
			GROUP BY category ORDER BY COUNT(*) DESC, category ASC`),

		"openings_by_location": groupCount(`
			SELECT location, COUNT(*) FROM jobs
			WHERE location != ''
			GROUP BY location ORDER BY COUNT(*) DESC, location ASC LIMIT 8`),

		"signups_by_month": groupCount(`
			SELECT substr(created_at, 1, 7), COUNT(*)
			FROM users GROUP BY substr(created_at, 1, 7)
			ORDER BY substr(created_at, 1, 7) ASC LIMIT 12`),
	})
}
