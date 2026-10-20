package main

import (
	"log"
	"net/http"
)

const addr = "localhost:8080"

func main() {
	openDB("alumnisetu.db")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", handleHealth)

	mux.HandleFunc("POST /api/register", handleRegister)
	mux.HandleFunc("POST /api/login", handleLogin)
	mux.HandleFunc("POST /api/logout", requireAuth(handleLogout))
	mux.HandleFunc("GET /api/me", requireAuth(handleMe))

	mux.HandleFunc("PUT /api/me/profile", requireAuth(handleUpdateProfile))
	mux.HandleFunc("GET /api/users/{id}", requireAuth(handleGetUser))

	mux.HandleFunc("GET /api/directory", requireAuth(handleDirectory))
	mux.HandleFunc("GET /api/directory/filters", requireAuth(handleFilterOptions))

	mux.HandleFunc("GET /api/recommendations", requireAuth(handleRecommendations))

	mux.HandleFunc("POST /api/mentorships", requireRole([]string{"student"}, handleCreateMentorship))
	mux.HandleFunc("GET /api/mentorships", requireAuth(handleListMentorships))
	mux.HandleFunc("PATCH /api/mentorships/{id}", requireRole([]string{"alumni"}, handleUpdateMentorship))

	log.Println("AlumniSetu API listening on http://" + addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "AlumniSetu API",
		"time":    now(),
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
