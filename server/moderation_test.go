package main

import "testing"

func TestModerate(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		flagged bool
	}{
		{"clean text", "Happy to help juniors with Go.", "Happy to help juniors with Go.", false},
		{"single bad word", "This advice is stupid.", "This advice is ******.", true},
		{"case insensitive", "STUPID question", "****** question", true},
		{"multi word phrase", "I offer a guaranteed job", "I offer a **************", true},
		{"substring is safe", "Scammers study Scandinavian design", "Scammers study Scandinavian design", false},
		{"keeps newlines", "line one\nstupid\nline three", "line one\n******\nline three", true},
		{"link spam", "a http://a.com http://b.com http://c.com http://d.com", "a http://a.com http://b.com http://c.com http://d.com", true},
		{"few links fine", "see http://a.com and http://b.com", "see http://a.com and http://b.com", false},
		{"shouting", "PLEASE CONTACT ME RIGHT NOW URGENT", "PLEASE CONTACT ME RIGHT NOW URGENT", true},
		{"short caps fine", "I WORK AT IBM", "I WORK AT IBM", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, flagged := moderate(c.input)
			if got != c.want {
				t.Errorf("text: got %q, want %q", got, c.want)
			}
			if flagged != c.flagged {
				t.Errorf("flagged: got %v, want %v", flagged, c.flagged)
			}
		})
	}
}
