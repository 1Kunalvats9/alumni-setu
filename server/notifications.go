package main

import (
	"net/http"
	"strconv"
)

func notify(userID int64, kind, title, body, page string) {
	db.Exec(`INSERT INTO notifications (user_id, kind, title, body, page, seen, created_at)
		VALUES (?, ?, ?, ?, ?, 0, ?)`, userID, kind, title, body, page, now())
}

func notifyEveryoneExcept(exceptUserID int64, kind, title, body, page string) {
	db.Exec(`INSERT INTO notifications (user_id, kind, title, body, page, seen, created_at)
		SELECT id, ?, ?, ?, ?, 0, ? FROM users WHERE id != ?`,
		kind, title, body, page, now(), exceptUserID)
}

func handleListNotifications(w http.ResponseWriter, r *http.Request, user *User) {
	rows, err := db.Query(`SELECT id, user_id, kind, title, body, page, seen, created_at
		FROM notifications WHERE user_id = ?
		ORDER BY created_at DESC, id DESC LIMIT 50`, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load notifications")
		return
	}
	defer rows.Close()

	list := []Notification{}
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Kind, &n.Title, &n.Body,
			&n.Page, &n.Seen, &n.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "could not read notifications")
			return
		}
		list = append(list, n)
	}

	unseen := countRows(`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND seen = 0`, user.ID)

	writeJSON(w, http.StatusOK, map[string]any{
		"count":   len(list),
		"unseen":  unseen,
		"results": list,
	})
}

func handleMarkAllSeen(w http.ResponseWriter, r *http.Request, user *User) {
	db.Exec(`UPDATE notifications SET seen = 1 WHERE user_id = ?`, user.ID)
	writeJSON(w, http.StatusOK, map[string]string{"status": "all marked as read"})
}

func handleDeleteNotification(w http.ResponseWriter, r *http.Request, user *User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	result, err := db.Exec(`DELETE FROM notifications WHERE id = ? AND user_id = ?`, id, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not remove notification")
		return
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		writeError(w, http.StatusNotFound, "no such notification")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}
