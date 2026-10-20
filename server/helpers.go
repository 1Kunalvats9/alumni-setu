package main

// helpers.go — tiny functions used by every handler, so the handlers
// themselves stay short and readable.

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// writeJSON sends any Go value back as a JSON response.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

// writeError sends a JSON error, e.g. {"error": "email already registered"}.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// readJSON decodes the request body into dest. It returns false and already
// writes the error response if the body is not valid JSON, so callers can do:
//
//	if !readJSON(w, r, &body) { return }
func readJSON(w http.ResponseWriter, r *http.Request, dest any) bool {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON in request body")
		return false
	}
	return true
}

// now returns the current time as a sortable text timestamp. SQLite has no
// real date type, so we store every timestamp as RFC3339 text.
func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// splitList turns "Go, React , ML" into ["go", "react", "ml"].
// Lower-cased and trimmed so comparisons are forgiving about how people type.
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

// sharedItems returns the entries that appear in both comma separated lists.
// Matching ignores case, but the result keeps the casing from `theirs`, so the
// UI shows "Machine Learning" rather than "machine learning".
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
