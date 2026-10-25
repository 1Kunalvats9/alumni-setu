package main

import (
	"net/http"
	"strconv"
	"strings"
)

var postCategories = []string{"general", "career", "academic", "placement"}

type postRequest struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Category string `json:"category"`
}

func handleCreatePost(w http.ResponseWriter, r *http.Request, user *User) {
	var body postRequest
	if !readJSON(w, r, &body) {
		return
	}

	body.Title = strings.TrimSpace(body.Title)
	body.Body = strings.TrimSpace(body.Body)

	if body.Title == "" || body.Body == "" {
		writeError(w, http.StatusBadRequest, "title and body are required")
		return
	}
	if body.Category == "" {
		body.Category = "general"
	}
	if !contains(postCategories, body.Category) {
		writeError(w, http.StatusBadRequest, "category must be general, career, academic or placement")
		return
	}

	cleanTitle, titleFlagged := moderate(body.Title)
	cleanBody, bodyFlagged := moderate(body.Body)
	flagged := titleFlagged || bodyFlagged

	result, err := db.Exec(`INSERT INTO posts (author_id, title, body, category, flagged, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		user.ID, cleanTitle, cleanBody, body.Category, flagged, now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create post")
		return
	}

	id, _ := result.LastInsertId()
	post, err := loadPost(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load post")
		return
	}
	writeJSON(w, http.StatusCreated, post)
}

func handleListPosts(w http.ResponseWriter, r *http.Request, _ *User) {
	where := []string{"1 = 1"}
	args := []any{}

	if category := r.URL.Query().Get("category"); category != "" {
		where = append(where, "p.category = ?")
		args = append(args, category)
	}
	if q := strings.TrimSpace(r.URL.Query().Get("q")); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		where = append(where, "(LOWER(p.title) LIKE ? OR LOWER(p.body) LIKE ?)")
		args = append(args, like, like)
	}

	rows, err := db.Query(postSelect+` WHERE `+strings.Join(where, " AND ")+
		` ORDER BY p.created_at DESC LIMIT 100`, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load posts")
		return
	}
	defer rows.Close()

	list := []Post{}
	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not read posts")
			return
		}
		list = append(list, post)
	}

	writeJSON(w, http.StatusOK, map[string]any{"count": len(list), "results": list})
}

func handleGetPost(w http.ResponseWriter, r *http.Request, _ *User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	post, err := loadPost(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "no such post")
		return
	}

	comments, err := loadComments(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load comments")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"post": post, "comments": comments})
}

type commentRequest struct {
	Body string `json:"body"`
}

func handleCreateComment(w http.ResponseWriter, r *http.Request, user *User) {
	postID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}
	if _, err := loadPost(postID); err != nil {
		writeError(w, http.StatusNotFound, "no such post")
		return
	}

	var body commentRequest
	if !readJSON(w, r, &body) {
		return
	}
	body.Body = strings.TrimSpace(body.Body)
	if body.Body == "" {
		writeError(w, http.StatusBadRequest, "comment cannot be empty")
		return
	}

	cleanBody, flagged := moderate(body.Body)

	result, err := db.Exec(`INSERT INTO comments (post_id, author_id, body, flagged, created_at)
		VALUES (?, ?, ?, ?, ?)`, postID, user.ID, cleanBody, flagged, now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not add comment")
		return
	}

	id, _ := result.LastInsertId()
	comment, err := scanComment(db.QueryRow(commentSelect+` WHERE c.id = ?`, id))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load comment")
		return
	}

	post, _ := loadPost(postID)
	if post.AuthorID != user.ID {
		notify(post.AuthorID, "reply",
			user.Name+" replied to \""+post.Title+"\"",
			comment.Body, "forum")
	}

	writeJSON(w, http.StatusCreated, comment)
}

func handleDeletePost(w http.ResponseWriter, r *http.Request, user *User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	post, err := loadPost(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "no such post")
		return
	}
	if post.AuthorID != user.ID && user.Role != "admin" {
		writeError(w, http.StatusForbidden, "you can only delete your own posts")
		return
	}

	db.Exec(`DELETE FROM posts WHERE id = ?`, id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func handleDeleteComment(w http.ResponseWriter, r *http.Request, user *User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid comment id")
		return
	}

	comment, err := scanComment(db.QueryRow(commentSelect+` WHERE c.id = ?`, id))
	if err != nil {
		writeError(w, http.StatusNotFound, "no such comment")
		return
	}
	if comment.AuthorID != user.ID && user.Role != "admin" {
		writeError(w, http.StatusForbidden, "you can only delete your own comments")
		return
	}

	db.Exec(`DELETE FROM comments WHERE id = ?`, id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

const postSelect = `
	SELECT p.id, p.author_id, author.name, author.role, p.title, p.body,
	       p.category, p.flagged, p.created_at,
	       (SELECT COUNT(*) FROM comments WHERE post_id = p.id)
	FROM posts p
	JOIN users author ON author.id = p.author_id`

const commentSelect = `
	SELECT c.id, c.post_id, c.author_id, author.name, author.role,
	       c.body, c.flagged, c.created_at
	FROM comments c
	JOIN users author ON author.id = c.author_id`

func scanPost(row rowScanner) (Post, error) {
	var p Post
	err := row.Scan(&p.ID, &p.AuthorID, &p.AuthorName, &p.AuthorRole,
		&p.Title, &p.Body, &p.Category, &p.Flagged, &p.CreatedAt, &p.CommentCount)
	return p, err
}

func scanComment(row rowScanner) (Comment, error) {
	var c Comment
	err := row.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.AuthorName,
		&c.AuthorRole, &c.Body, &c.Flagged, &c.CreatedAt)
	return c, err
}

func loadPost(id int64) (Post, error) {
	return scanPost(db.QueryRow(postSelect+` WHERE p.id = ?`, id))
}

func loadComments(postID int64) ([]Comment, error) {
	rows, err := db.Query(commentSelect+` WHERE c.post_id = ? ORDER BY c.created_at ASC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Comment{}
	for rows.Next() {
		comment, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, comment)
	}
	return list, nil
}
