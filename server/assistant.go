package main

import (
	"net/http"
	"strconv"
	"strings"
)

type intent struct {
	Name        string
	Keywords    []string
	Text        string
	Live        func(user *User) string
	Page        string
	Suggestions []string
}

var intents = []intent{
	{
		Name:     "find_mentor",
		Keywords: []string{"mentor", "guidance", "guide", "senior", "advice", "career help"},
		Text: "Open Suggested Mentors. Alumni are ranked by how well they match your " +
			"profile, and each card lists the reasons behind the match. Click View " +
			"profile on anyone and send them a request with a short message about " +
			"what you need help with.",
		Page:        "mentors",
		Suggestions: []string{"How is the match score calculated?", "Why do I have no suggestions?"},
	},
	{
		Name:     "match_score",
		Keywords: []string{"score", "match", "matched", "ranked", "ranking", "recommend", "suggestion", "algorithm"},
		Text: "The score adds up what you have in common: 10 points for each shared " +
			"skill, 6 for each shared interest, 15 for the same branch, 8 if they work " +
			"in an industry you follow, 5 for the same college, 12 if they are open to " +
			"mentoring and 5 if the department has verified them. Every suggestion " +
			"lists the reasons that earned it those points.",
		Page:        "mentors",
		Suggestions: []string{"How do I get better matches?", "What does verified mean?"},
	},
	{
		Name:     "improve_matches",
		Keywords: []string{"no suggestions", "better match", "improve", "empty", "nothing showing", "no results"},
		Text: "Matching reads your profile, so fill it in. Branch, skills and interests " +
			"matter most. Open My Profile and add your skills as a comma separated " +
			"list, for example 'Python, SQL, React'.",
		Page:        "profile",
		Suggestions: []string{"How do I update my profile?"},
	},
	{
		Name:     "update_profile",
		Keywords: []string{"profile", "update", "edit", "change my", "details", "bio", "skills"},
		Text: "Open My Profile. Everything there appears in the directory and feeds " +
			"mentor matching. Skills and interests are comma separated. Alumni can " +
			"also tick 'open to mentoring' so students can find them.",
		Page:        "profile",
		Suggestions: []string{"Who can see my profile?", "What does verified mean?"},
	},
	{
		Name:     "verification",
		Keywords: []string{"verified", "verify", "verification", "fake", "genuine", "authentic", "trust"},
		Text: "A verified badge means the Technical Education Department has checked " +
			"that person's degree and employment. Verified profiles rank higher in " +
			"matching, and the directory can be filtered to show only verified people. " +
			"New alumni accounts start unverified until an admin reviews them.",
		Suggestions: []string{"How long does verification take?", "How do I report a fake profile?"},
	},
	{
		Name:     "report_abuse",
		Keywords: []string{"report", "abuse", "abusive", "offensive", "harassment", "spam", "inappropriate", "fake profile"},
		Text: "Every post, reply, bio and job description passes through an automatic " +
			"filter before it is saved. Offensive words are masked and the content is " +
			"flagged for an admin to review in the Admin Console. If something slipped " +
			"through, contact the department admin.",
		Suggestions: []string{"What does the filter block?"},
	},
	{
		Name:     "events",
		Keywords: []string{"event", "events", "webinar", "session", "meetup", "panel", "workshop", "rsvp", "attend"},
		Live: func(user *User) string {
			var title, startsAt string
			err := db.QueryRow(`SELECT title, starts_at FROM events
				WHERE starts_at >= ? ORDER BY starts_at ASC LIMIT 1`, now()).Scan(&title, &startsAt)
			if err != nil {
				return "There are no upcoming sessions scheduled right now. " +
					"Open Events & Webinars to see past sessions."
			}
			return "The next session is \"" + title + "\" on " + prettyTime(startsAt) +
				". Open Events & Webinars and click Attend to RSVP. " +
				"Alumni and admins can also schedule new sessions from that page."
		},
		Page:        "events",
		Suggestions: []string{"Can I host a session?"},
	},
	{
		Name: "host_event",
		Keywords: []string{"host", "host a session", "host a webinar", "schedule a",
			"schedule a session", "create event", "organise", "organize", "conduct"},
		Text: "Alumni and department admins can schedule sessions. Open Events & " +
			"Webinars and click 'Schedule a session'. You can pick a webinar, an " +
			"alumni meetup or a panel discussion, and add a joining link or venue. " +
			"Students can attend but cannot create sessions.",
		Page: "events",
	},
	{
		Name:     "jobs",
		Keywords: []string{"job", "jobs", "internship", "placement", "vacancy", "opening", "hiring", "apply"},
		Live: func(user *User) string {
			count := countRows(`SELECT COUNT(*) FROM jobs`)
			interns := countRows(`SELECT COUNT(*) FROM jobs WHERE kind = 'internship'`)
			return "There are " + strconv.Itoa(count) + " openings on the Placements " +
				"board right now, " + strconv.Itoa(interns) + " of them internships. " +
				"They are posted by alumni, and each one carries an application link. " +
				"Alumni can post openings from the same page."
		},
		Page:        "jobs",
		Suggestions: []string{"How do I ask an alumnus for a referral?"},
	},
	{
		Name:     "referral",
		Keywords: []string{"referral", "refer", "recommendation letter", "reference"},
		Text: "Find the person in the Alumni Directory, filter by their company, and " +
			"send a mentorship request introducing yourself. Mention the specific role " +
			"and attach your details. A referral is much more likely after a real " +
			"conversation than as a first message.",
		Page: "directory",
	},
	{
		Name:     "forum",
		Keywords: []string{"forum", "discussion", "question", "ask", "post", "thread", "reply", "doubt"},
		Text: "Open Discussion Forum and click 'Start a discussion'. Pick a topic: " +
			"general, career, academic or placement. Alumni and students both answer " +
			"there, and you can filter the list by topic to find existing answers " +
			"before you ask.",
		Page:        "forum",
		Suggestions: []string{"What does the filter block?"},
	},
	{
		Name: "directory_search",
		Keywords: []string{"directory", "search", "find alumni", "filter", "batch",
			"branch", "company", "who works", "how many alumni", "how many people"},
		Live: func(user *User) string {
			alumni := countRows(`SELECT COUNT(*) FROM users WHERE role = 'alumni'`)
			return "The Alumni Directory holds " + strconv.Itoa(alumni) + " alumni. " +
				"Search by name, company, skill, branch or industry, and narrow the " +
				"list with the branch, batch and 'open to mentoring' filters. Click " +
				"any card to see the full profile."
		},
		Page: "directory",
	},
	{
		Name:     "mentorship_status",
		Keywords: []string{"my request", "pending", "accepted", "rejected", "status", "replied"},
		Live: func(user *User) string {
			pending := countRows(`SELECT COUNT(*) FROM mentorships
				WHERE (student_id = ? OR alumni_id = ?) AND status = 'pending'`, user.ID, user.ID)
			accepted := countRows(`SELECT COUNT(*) FROM mentorships
				WHERE (student_id = ? OR alumni_id = ?) AND status = 'accepted'`, user.ID, user.ID)
			return "You have " + strconv.Itoa(pending) + " pending and " +
				strconv.Itoa(accepted) + " accepted mentorship connections. " +
				"Open Mentorship Requests to see them. Mentors accept or decline from " +
				"that same page."
		},
		Page: "mentorships",
	},
	{
		Name:     "filter_rules",
		Keywords: []string{"filter block", "masked", "asterisk", "stars", "censored", "flagged"},
		Text: "The filter masks abusive words and common scam phrases with asterisks, " +
			"and flags anything with more than three links or written mostly in " +
			"capitals. Word boundaries are respected, so a legitimate word that merely " +
			"contains a blocked word is left alone. Flagged content stays visible but " +
			"is queued for an admin.",
	},
	{
		Name:     "privacy",
		Keywords: []string{"who can see", "privacy", "private", "visible", "data", "secure", "password"},
		Text: "Anyone signed in can see your name, role and profile details in the " +
			"directory. Your password is never stored, only a bcrypt hash of it. " +
			"Your login token can be revoked at any time by signing out.",
		Page: "profile",
	},
}

const fallbackAnswer = "I can help with finding a mentor, how match scores are " +
	"calculated, updating your profile, verification, events and webinars, the " +
	"placements board, the discussion forum, searching the alumni directory, and " +
	"reporting inappropriate content. Try asking about one of those."

var fallbackSuggestions = []string{
	"How do I find a mentor?",
	"When is the next session?",
	"What jobs are available?",
	"How do I update my profile?",
}

func bestIntent(question string) *intent {
	text := strings.ToLower(question)

	best := -1
	bestScore := 0

	for i, candidate := range intents {
		score := 0
		for _, keyword := range candidate.Keywords {
			if strings.Contains(text, keyword) {
				score += len(strings.Fields(keyword))
			}
		}
		if score > bestScore {
			best, bestScore = i, score
		}
	}

	if best == -1 {
		return nil
	}
	return &intents[best]
}

type assistantRequest struct {
	Question string `json:"question"`
}

func handleAssistant(w http.ResponseWriter, r *http.Request, user *User) {
	var body assistantRequest
	if !readJSON(w, r, &body) {
		return
	}

	body.Question = strings.TrimSpace(body.Question)
	if body.Question == "" {
		writeError(w, http.StatusBadRequest, "please type a question")
		return
	}

	match := bestIntent(body.Question)
	if match == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"answer":      fallbackAnswer,
			"matched":     "",
			"page":        "",
			"suggestions": fallbackSuggestions,
		})
		return
	}

	answer := match.Text
	if match.Live != nil {
		answer = match.Live(user)
	}

	suggestions := match.Suggestions
	if len(suggestions) == 0 {
		suggestions = fallbackSuggestions[:2]
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"answer":      answer,
		"matched":     match.Name,
		"page":        match.Page,
		"suggestions": suggestions,
	})
}
