package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func readJSON(w http.ResponseWriter, r *http.Request, dest any) bool {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON in request body")
		return false
	}
	return true
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func prettyTime(value string) string {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return parsed.Format("2 Jan 2006 at 15:04 UTC")
}

func splitList(csv string) []string {
	var out []string
	for _, part := range strings.Split(csv, ",") {
		part = strings.ToLower(strings.TrimSpace(part))
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func sharedItems(mine, theirs string) []string {
	have := make(map[string]bool)
	for _, item := range splitList(mine) {
		have[item] = true
	}

	var shared []string
	for _, original := range strings.Split(theirs, ",") {
		original = strings.TrimSpace(original)
		if original != "" && have[strings.ToLower(original)] {
			shared = append(shared, original)
		}
	}
	return shared
}

func tidyList(csv string) string {
	var parts []string
	for _, part := range strings.Split(csv, ",") {
		if part = strings.TrimSpace(part); part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ", ")
}
