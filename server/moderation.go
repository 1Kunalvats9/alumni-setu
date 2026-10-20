package main

import (
	"regexp"
	"strings"
)

var blockedWords = []string{
	"abuse", "bastard", "bitch", "bloody", "crap", "damn", "dumb",
	"fuck", "idiot", "loser", "moron", "nonsense", "rubbish",
	"shit", "stupid", "useless", "worthless",
	"scam", "fraud", "fake degree", "buy followers", "guaranteed job",
	"pay me", "send money", "quick money", "lottery", "crypto giveaway",
}

var blockedPattern = buildBlockedPattern()

var linkPattern = regexp.MustCompile(`(?i)https?://\S+`)

func buildBlockedPattern() *regexp.Regexp {
	escaped := make([]string, len(blockedWords))
	for i, word := range blockedWords {
		escaped[i] = regexp.QuoteMeta(word)
	}
	return regexp.MustCompile(`(?i)\b(` + strings.Join(escaped, "|") + `)\b`)
}

func moderate(text string) (string, bool) {
	flagged := false

	cleaned := blockedPattern.ReplaceAllStringFunc(text, func(match string) string {
		flagged = true
		return strings.Repeat("*", len([]rune(match)))
	})

	if len(linkPattern.FindAllString(text, -1)) > 3 {
		flagged = true
	}

	if isShouting(text) {
		flagged = true
	}

	return cleaned, flagged
}

func isShouting(text string) bool {
	letters, upper := 0, 0
	for _, r := range text {
		if r >= 'a' && r <= 'z' {
			letters++
		}
		if r >= 'A' && r <= 'Z' {
			letters++
			upper++
		}
	}
	return letters >= 20 && upper*100/letters > 70
}
