package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func newToken() string {
	raw := make([]byte, 32)
	rand.Read(raw)
	return hex.EncodeToString(raw)
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var body registerRequest
	if !readJSON(w, r, &body) {
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))

	if body.Name == "" || body.Email == "" {
		writeError(w, http.StatusBadRequest, "name and email are required")
		return
	}
	if len(body.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if body.Role != "student" && body.Role != "alumni" {
		writeError(w, http.StatusBadRequest, "role must be 'student' or 'alumni'")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not secure password")
		return
	}

	result, err := db.Exec(
		`INSERT INTO users (name, email, password_hash, role, verified, created_at)
		 VALUES (?, ?, ?, ?, 0, ?)`,
		body.Name, body.Email, string(hash), body.Role, now(),
	)
	if err != nil {
		writeError(w, http.StatusConflict, "that email is already registered")
		return
	}

	userID, _ := result.LastInsertId()

	db.Exec(`INSERT INTO profiles (user_id, updated_at) VALUES (?, ?)`, userID, now())

	token := newToken()
	db.Exec(`INSERT INTO sessions (token, user_id, created_at) VALUES (?, ?, ?)`,
		token, userID, now())

	user, _ := findUserByID(userID)
	writeJSON(w, http.StatusCreated, map[string]any{"token": token, "user": user})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if !readJSON(w, r, &body) {
		return
	}
	email := strings.ToLower(strings.TrimSpace(body.Email))

	var user User
	err := db.QueryRow(
		`SELECT id, name, email, password_hash, role, verified, created_at
		 FROM users WHERE email = ?`, email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password,
		&user.Role, &user.Verified, &user.CreatedAt)

	if err == sql.ErrNoRows {
		writeError(w, http.StatusUnauthorized, "incorrect email or password")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "incorrect email or password")
		return
	}

	token := newToken()
	db.Exec(`INSERT INTO sessions (token, user_id, created_at) VALUES (?, ?, ?)`,
		token, user.ID, now())

	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": user})
}

func handleLogout(w http.ResponseWriter, r *http.Request, _ *User) {
	db.Exec(`DELETE FROM sessions WHERE token = ?`, bearerToken(r))
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}

func findUserByID(id int64) (*User, error) {
	var u User
	err := db.QueryRow(
		`SELECT id, name, email, role, verified, created_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Verified, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func userFromRequest(r *http.Request) (*User, error) {
	var u User
	err := db.QueryRow(
		`SELECT u.id, u.name, u.email, u.role, u.verified, u.created_at
		 FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.token = ?`, bearerToken(r),
	).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Verified, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

type authedHandler func(w http.ResponseWriter, r *http.Request, user *User)

func requireAuth(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := userFromRequest(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "please log in")
			return
		}
		next(w, r, user)
	}
}

func requireRole(roles []string, next authedHandler) http.HandlerFunc {
	return requireAuth(func(w http.ResponseWriter, r *http.Request, user *User) {
		for _, role := range roles {
			if user.Role == role {
				next(w, r, user)
				return
			}
		}
		writeError(w, http.StatusForbidden, "your account type cannot do that")
	})
}

func handleMe(w http.ResponseWriter, r *http.Request, user *User) {
	profile, err := loadProfile(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load profile")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user, "profile": profile})
}
