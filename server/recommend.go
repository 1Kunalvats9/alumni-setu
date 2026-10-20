package main

// recommend.go — suggests who a user should connect with.
//
// The problem statement asks the platform to "suggest connections between
// students and alumni with similar interests, career paths, industry and
// skills". This file does exactly that with a transparent, weighted
// overlap score rather than a trained model.
//
// Why scoring and not machine learning? Three reasons:
//   1. It works from day one, with no training data. A new platform has none.
//   2. Every suggestion can be explained: "3 shared skills: Go, React, SQL".
//      A student trusts a reason far more than a number out of a black box.
//   3. The weights below are the only thing an ML model would eventually
//      replace, so this is a drop-in place to upgrade later.

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// These weights decide what "a good match" means. They are deliberately in
// one place so they are easy to find, explain and tune.
const (
	pointsPerSharedSkill    = 10 // strongest signal: they do the same work
	pointsPerSharedInterest = 6
	pointsSameBranch        = 15 // studied the same subject
	pointsIndustryMatch     = 8  // the student wants that industry
	pointsSameCollege       = 5
	pointsVerified          = 5 // nudge trusted profiles up
	pointsOpenToMentor      = 12
)

// handleRecommendations returns the best matches for the logged-in user.
// Students are matched with alumni, alumni are matched with students.
func handleRecommendations(w http.ResponseWriter, r *http.Request, user *User) {
	limit := 5
	if n, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && n > 0 && n <= 50 {
		limit = n
	}

	me, err := loadProfile(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load your profile")
		return
	}

	// Look at the other side of the platform: a student wants alumni, an
	// alumnus wants students to mentor. Admins see alumni.
	targetRole := "alumni"
	if user.Role == "alumni" {
		targetRole = "student"
	}

	candidates, err := candidatesWithRole(targetRole, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load candidates")
		return
	}

	// Score every candidate, then keep the ones that match on something.
	recommendations := []Recommendation{}
	for _, candidate := range candidates {
		score, reasons := matchScore(me, candidate.Profile, candidate.Verified)
		if score == 0 {
			continue // nothing in common, do not suggest them at all
		}
		recommendations = append(recommendations, Recommendation{
			DirectoryEntry: candidate,
			Score:          score,
			Reasons:        reasons,
		})
	}

	// Highest score first; ties broken by name so the order is stable.
	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].Score != recommendations[j].Score {
			return recommendations[i].Score > recommendations[j].Score
		}
		return recommendations[i].Name < recommendations[j].Name
	})

	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"count":   len(recommendations),
		"results": recommendations,
	})
}

// matchScore compares the viewer's profile against one candidate and returns
// the score plus a plain-English reason for each point awarded.
func matchScore(me, them Profile, theyAreVerified bool) (int, []string) {
	score := 0
	reasons := []string{}

	// 1. Shared skills — the single strongest signal.
	if shared := sharedItems(me.Skills, them.Skills); len(shared) > 0 {
		score += len(shared) * pointsPerSharedSkill
		reasons = append(reasons, label(len(shared), "shared skill")+": "+strings.Join(shared, ", "))
	}

	// 2. Shared interests.
	if shared := sharedItems(me.Interests, them.Interests); len(shared) > 0 {
		score += len(shared) * pointsPerSharedInterest
		reasons = append(reasons, label(len(shared), "shared interest")+": "+strings.Join(shared, ", "))
	}

	// 3. Same branch of engineering.
	if me.Branch != "" && strings.EqualFold(me.Branch, them.Branch) {
		score += pointsSameBranch
		reasons = append(reasons, "Same branch: "+them.Branch)
	}

	// 4. The candidate works in an industry the viewer listed as an interest.
	if them.Industry != "" {
		for _, interest := range splitList(me.Interests) {
			if interest == strings.ToLower(them.Industry) {
				score += pointsIndustryMatch
				reasons = append(reasons, "Works in "+them.Industry+", which you follow")
				break
			}
		}
	}

	// 5. Same college.
	if me.College != "" && strings.EqualFold(me.College, them.College) {
		score += pointsSameCollege
		reasons = append(reasons, "Same college: "+them.College)
	}

	// 6. Willing to mentor.
	if them.OpenToMentor {
		score += pointsOpenToMentor
		reasons = append(reasons, "Open to mentoring")
	}

	// 7. Verified by the department.
	if theyAreVerified {
		score += pointsVerified
		reasons = append(reasons, "Verified profile")
	}

	return score, reasons
}

// label returns "1 shared skill" or "3 shared skills".
func label(n int, word string) string {
	text := strconv.Itoa(n) + " " + word
	if n != 1 {
		text += "s"
	}
	return text
}

// candidatesWithRole loads everyone of a given role, except the viewer.
func candidatesWithRole(role string, excludeUserID int64) ([]DirectoryEntry, error) {
	rows, err := db.Query(`
		SELECT u.id, u.name, u.email, u.role, u.verified, u.created_at, `+profileColumns+`
		FROM users u
		JOIN profiles p ON p.user_id = u.id
		WHERE u.role = ? AND u.id != ?`, role, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
			return nil, err
		}
		e.Profile = p
		entries = append(entries, e)
	}
	return entries, nil
}
