package main

// main.go — the entry point. It opens the database, registers every URL the
// API answers, and starts listening.
//
// Everything is standard library net/http. A "route" is just a path string
// mapped to a function, and Go 1.22+ lets us put the HTTP method in the
// pattern itself ("GET /api/health"), which keeps the table easy to read.

import (
	"log"
	"net/http"
)

const addr = "localhost:8080"

func main() {
	openDB("alumnisetu.db")

	mux := http.NewServeMux()

	// --- Health ----------------------------------------------------------
	mux.HandleFunc("GET /api/health", handleHealth)

	log.Println("AlumniSetu API listening on http://" + addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

// handleHealth is a tiny endpoint the desktop app pings on startup to check
// that the server is up before it shows the login screen.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "AlumniSetu API",
		"time":    now(),
	})
}

// withCORS wraps the router so the Electron window is allowed to call the API.
// The renderer page is loaded from a file:// URL, so the browser treats every
// request to localhost:8080 as cross-origin and would block it without these
// headers.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")

		// Browsers send an OPTIONS "preflight" before the real request.
		// Answer it immediately with no body.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
