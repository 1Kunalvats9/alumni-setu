package main

import (
	"log"
	"net/http"
	"os"
)

const defaultAddr = "localhost:8080"

func setting(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func main() {
	addr := setting("ALUMNISETU_ADDR", defaultAddr)

	openDB(setting("ALUMNISETU_DB", "alumnisetu.db"))
	seedIfEmpty()

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

	mux.HandleFunc("GET /api/events", requireAuth(handleListEvents))
	mux.HandleFunc("POST /api/events", requireRole([]string{"alumni", "admin"}, handleCreateEvent))
	mux.HandleFunc("POST /api/events/{id}/rsvp", requireAuth(handleRSVP))
	mux.HandleFunc("DELETE /api/events/{id}/rsvp", requireAuth(handleRSVP))

	mux.HandleFunc("GET /api/posts", requireAuth(handleListPosts))
	mux.HandleFunc("POST /api/posts", requireAuth(handleCreatePost))
	mux.HandleFunc("GET /api/posts/{id}", requireAuth(handleGetPost))
	mux.HandleFunc("DELETE /api/posts/{id}", requireAuth(handleDeletePost))
	mux.HandleFunc("POST /api/posts/{id}/comments", requireAuth(handleCreateComment))
	mux.HandleFunc("DELETE /api/comments/{id}", requireAuth(handleDeleteComment))

	mux.HandleFunc("GET /api/jobs", requireAuth(handleListJobs))
	mux.HandleFunc("POST /api/jobs", requireRole([]string{"alumni", "admin"}, handleCreateJob))
	mux.HandleFunc("DELETE /api/jobs/{id}", requireAuth(handleDeleteJob))

	mux.HandleFunc("GET /api/notifications", requireAuth(handleListNotifications))
	mux.HandleFunc("POST /api/notifications/read", requireAuth(handleMarkAllSeen))
	mux.HandleFunc("DELETE /api/notifications/{id}", requireAuth(handleDeleteNotification))

	mux.HandleFunc("POST /api/me/extract-skills", requireAuth(handleExtractSkills))

	mux.HandleFunc("POST /api/assistant", requireAuth(handleAssistant))

	admin := []string{"admin"}
	mux.HandleFunc("GET /api/admin/stats", requireRole(admin, handleAdminStats))
	mux.HandleFunc("GET /api/admin/pending", requireRole(admin, handleAdminPending))
	mux.HandleFunc("PATCH /api/admin/users/{id}/verify", requireRole(admin, handleAdminVerify))
	mux.HandleFunc("GET /api/admin/flagged", requireRole(admin, handleAdminFlagged))
	mux.HandleFunc("GET /api/admin/ledger", requireRole(admin, handleLedger))
	mux.HandleFunc("GET /api/admin/analytics", requireRole(admin, handleAnalytics))
	mux.HandleFunc("PATCH /api/admin/posts/{id}/approve", requireRole(admin, handleAdminClearFlag))

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
