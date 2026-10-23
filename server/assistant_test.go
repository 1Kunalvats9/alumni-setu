package main

import "testing"

func TestBestIntent(t *testing.T) {
	cases := []struct {
		question string
		want     string
	}{
		{"How do I find a mentor?", "find_mentor"},
		{"who can guide me on my career", "find_mentor"},
		{"how is the match score calculated", "match_score"},
		{"why do I have no suggestions", "improve_matches"},
		{"how do I update my profile", "update_profile"},
		{"what does verified mean", "verification"},
		{"someone posted something offensive", "report_abuse"},
		{"when is the next webinar", "events"},
		{"can I host a session", "host_event"},
		{"are there any internships", "jobs"},
		{"how do I ask for a referral", "referral"},
		{"where do I post a doubt", "forum"},
		{"how do I search the directory", "directory_search"},
		{"what is the weather in Jaipur", ""},
		{"", ""},
	}

	for _, c := range cases {
		t.Run(c.question, func(t *testing.T) {
			match := bestIntent(c.question)

			if c.want == "" {
				if match != nil {
					t.Errorf("expected no match, got %q", match.Name)
				}
				return
			}
			if match == nil {
				t.Fatalf("expected %q, got no match", c.want)
			}
			if match.Name != c.want {
				t.Errorf("got %q, want %q", match.Name, c.want)
			}
		})
	}
}
