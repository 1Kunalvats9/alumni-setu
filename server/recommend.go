package main

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
)

const (
	pointsPerSharedSkill    = 10
	pointsPerSharedInterest = 6
	pointsSameBranch        = 15
	pointsIndustryMatch     = 8
	pointsSameCollege       = 5
	pointsVerified          = 5
	pointsOpenToMentor      = 12
)

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

	targetRole := "alumni"
	if user.Role == "alumni" {
		targetRole = "student"
	}

	candidates, err := candidatesWithRole(targetRole, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load candidates")
		return
	}

	recommendations := []Recommendation{}
	for _, candidate := range candidates {
		score, reasons := matchScore(me, candidate.Profile, candidate.Verified)
		if score == 0 {
			continue
		}
		recommendations = append(recommendations, Recommendation{
			DirectoryEntry: candidate,
			Score:          score,
			Reasons:        reasons,
		})
	}

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

func matchScore(me, them Profile, theyAreVerified bool) (int, []string) {
	score := 0
	reasons := []string{}

	if shared := sharedItems(me.Skills, them.Skills); len(shared) > 0 {
		score += len(shared) * pointsPerSharedSkill
		reasons = append(reasons, label(len(shared), "shared skill")+": "+strings.Join(shared, ", "))
	}

	if shared := sharedItems(me.Interests, them.Interests); len(shared) > 0 {
		score += len(shared) * pointsPerSharedInterest
		reasons = append(reasons, label(len(shared), "shared interest")+": "+strings.Join(shared, ", "))
	}

	if me.Branch != "" && strings.EqualFold(me.Branch, them.Branch) {
		score += pointsSameBranch
		reasons = append(reasons, "Same branch: "+them.Branch)
	}

	if them.Industry != "" {
		for _, interest := range splitList(me.Interests) {
			if interest == strings.ToLower(them.Industry) {
				score += pointsIndustryMatch
				reasons = append(reasons, "Works in "+them.Industry+", which you follow")
				break
			}
		}
	}

	if me.College != "" && strings.EqualFold(me.College, them.College) {
		score += pointsSameCollege
		reasons = append(reasons, "Same college: "+them.College)
	}

	if them.OpenToMentor {
		score += pointsOpenToMentor
		reasons = append(reasons, "Open to mentoring")
	}

	if theyAreVerified {
		score += pointsVerified
		reasons = append(reasons, "Verified profile")
	}

	return score, reasons
}

func label(n int, word string) string {
	text := strconv.Itoa(n) + " " + word
	if n != 1 {
		text += "s"
	}
	return text
}

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
		entry, err := scanDirectoryRow(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
