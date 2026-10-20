package main

// auth.go — registration, login, logout and the "is this request logged in?"
// check that protects every other endpoint.
//
// How the login works, in three steps:
//  1. Register: we bcrypt the password and store only the hash.
//  2. Login: we compare the typed password against the hash. If it matches we
//     generate a long random token and save it in the sessions table.
//  3. Every later request sends "Authorization: Bearer <token>". We look the
//     token up to find out who is calling.
//
// This is the same idea as a JWT, but with the token stored in our own table,
// which makes it much easier to read and to revoke (logout = delete the row).

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// newToken returns 64 random hex characters, used as a session token.
func newToken() string {
	raw := make([]byte, 32)
	rand.Read(raw) // crypto/rand never fails on the platforms we target
	return hex.EncodeToString(raw)
}

// --- Registration --------------------------------------------------------

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // "student" or "alumni"
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var body registerRequest
	if !readJSON(w, r, &body) {
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))

	// Validate before touching the database.
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

	// bcrypt both hashes and salts the password for us.
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
		// The UNIQUE constraint on email is what usually fails here.
		writeError(w, http.StatusConflict, "that email is already registered")
		return
	}

	userID, _ := result.LastInsertId()

	// Give every new user an empty profile row so later updates are a simple
	// UPDATE and we never have to worry about whether the row exists.
	db.Exec(`INSERT INTO profiles (user_id, updated_at) VALUES (?, ?)`, userID, now())

	// Log them straight in so they do not have to type the password twice.
	token := newToken()
	db.Exec(`INSERT INTO sessions (token, user_id, created_at) VALUES (?, ?, ?)`,
		token, userID, now())

	user, _ := findUserByID(userID)
	writeJSON(w, http.StatusCreated, map[string]any{"token": token, "user": user})
}

// --- Login / logout ------------------------------------------------------

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

	// Deliberately give the same message for "no such email" and "wrong
	// password", so nobody can use the login form to discover which emails
	// are registered.
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

// handleLogout deletes the session row, which instantly invalidates the token.
func handleLogout(w http.ResponseWriter, r *http.Request, _ *User) {
	db.Exec(`DELETE FROM sessions WHERE token = ?`, bearerToken(r))
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// --- Checking who is calling ---------------------------------------------

// bearerToken pulls the token out of the "Authorization: Bearer xxx" header.
func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}

// findUserByID loads one user, without the password hash.
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

// userFromRequest joins sessions to users to find the caller.
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

// authedHandler is a normal handler that also receives the logged-in user, so
// handlers never have to repeat the lookup themselves.
type authedHandler func(w http.ResponseWriter, r *http.Request, user *User)

// requireAuth turns an authedHandler into a plain http.HandlerFunc, rejecting
// the request with 401 if the token is missing or unknown.
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

// requireRole is requireAuth plus a check that the user has one of the given
// roles, e.g. only alumni may post jobs.
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

// handleMe returns the logged-in user. The desktop app calls this on startup
// to check whether the saved token is still valid.
func handleMe(w http.ResponseWriter, r *http.Request, user *User) {
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}
