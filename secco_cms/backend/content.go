package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Content struct {
	ID          int64   `json:"id"`
	Type        string  `json:"type"`
	Slug        string  `json:"slug"`
	Title       string  `json:"title"`
	Summary     string  `json:"summary"`
	Body        string  `json:"body"`
	CoverImage  string  `json:"cover_image"`
	Status      string  `json:"status"`
	SortOrder   int     `json:"sort_order"`
	PublishedAt *string `json:"published_at"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type ContentInput struct {
	Type       string `json:"type"`
	Title      string `json:"title"`
	Slug       string `json:"slug"`
	Summary    string `json:"summary"`
	Body       string `json:"body"`
	CoverImage string `json:"cover_image"`
	SortOrder  int    `json:"sort_order"`
}

var validTypes = map[string]bool{
	"article": true,
	"project": true,
	"process": true,
}

func handleListContent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		query := `SELECT id, type, slug, title, summary, cover_image, status, sort_order, published_at, created_at, updated_at FROM content WHERE 1=1`
		var args []interface{}

		if t := r.URL.Query().Get("type"); t != "" {
			query += " AND type = ?"
			args = append(args, t)
		}
		if s := r.URL.Query().Get("status"); s != "" {
			query += " AND status = ?"
			args = append(args, s)
		}

		query += " ORDER BY type, sort_order, created_at DESC"

		rows, err := db.Query(query, args...)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to query content")
			return
		}
		defer rows.Close()

		contents := []Content{}
		for rows.Next() {
			var c Content
			if err := rows.Scan(&c.ID, &c.Type, &c.Slug, &c.Title, &c.Summary, &c.CoverImage, &c.Status, &c.SortOrder, &c.PublishedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to scan content")
				return
			}
			contents = append(contents, c)
		}

		json.NewEncoder(w).Encode(contents)
	}
}

func handleGetContent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "missing id")
			return
		}

		var c Content
		err := db.QueryRow(
			`SELECT id, type, slug, title, summary, body, cover_image, status, sort_order, published_at, created_at, updated_at FROM content WHERE id = ?`,
			id,
		).Scan(&c.ID, &c.Type, &c.Slug, &c.Title, &c.Summary, &c.Body, &c.CoverImage, &c.Status, &c.SortOrder, &c.PublishedAt, &c.CreatedAt, &c.UpdatedAt)
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "content not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to query content")
			return
		}

		json.NewEncoder(w).Encode(c)
	}
}

func handleCreateContent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var input ContentInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if strings.TrimSpace(input.Title) == "" {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}
		if !validTypes[input.Type] {
			writeError(w, http.StatusBadRequest, "type must be article, project, or process")
			return
		}

		// Process type: only one allowed
		if input.Type == "process" {
			var count int
			if err := db.QueryRow(`SELECT COUNT(*) FROM content WHERE type = 'process'`).Scan(&count); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to check process count")
				return
			}
			if count > 0 {
				writeError(w, http.StatusConflict, "only one process entry is allowed")
				return
			}
		}

		// Auto-generate slug if empty
		slug := strings.TrimSpace(input.Slug)
		if slug == "" {
			slug = slugify(input.Title)
		}

		// Validate slug uniqueness
		var exists int
		if err := db.QueryRow(`SELECT COUNT(*) FROM content WHERE slug = ?`, slug).Scan(&exists); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to check slug")
			return
		}
		if exists > 0 {
			writeError(w, http.StatusConflict, "slug already exists")
			return
		}

		result, err := db.Exec(
			`INSERT INTO content (type, slug, title, summary, body, cover_image, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			input.Type, slug, input.Title, input.Summary, input.Body, input.CoverImage, input.SortOrder,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create content")
			return
		}

		id, _ := result.LastInsertId()

		var c Content
		err = db.QueryRow(
			`SELECT id, type, slug, title, summary, body, cover_image, status, sort_order, published_at, created_at, updated_at FROM content WHERE id = ?`,
			id,
		).Scan(&c.ID, &c.Type, &c.Slug, &c.Title, &c.Summary, &c.Body, &c.CoverImage, &c.Status, &c.SortOrder, &c.PublishedAt, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read created content")
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(c)
	}
}

func handleUpdateContent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "missing id")
			return
		}

		// Check content exists
		var existingID int64
		if err := db.QueryRow(`SELECT id FROM content WHERE id = ?`, id).Scan(&existingID); err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "content not found")
			return
		}

		var input ContentInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Check slug uniqueness if changed
		slug := strings.TrimSpace(input.Slug)
		if slug == "" {
			slug = slugify(input.Title)
		}

		var slugCount int
		if err := db.QueryRow(`SELECT COUNT(*) FROM content WHERE slug = ? AND id != ?`, slug, id).Scan(&slugCount); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to check slug")
			return
		}
		if slugCount > 0 {
			writeError(w, http.StatusConflict, "slug already exists")
			return
		}

		_, err := db.Exec(
			`UPDATE content SET type = ?, slug = ?, title = ?, summary = ?, body = ?, cover_image = ?, sort_order = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			input.Type, slug, input.Title, input.Summary, input.Body, input.CoverImage, input.SortOrder, id,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update content")
			return
		}

		var c Content
		err = db.QueryRow(
			`SELECT id, type, slug, title, summary, body, cover_image, status, sort_order, published_at, created_at, updated_at FROM content WHERE id = ?`,
			id,
		).Scan(&c.ID, &c.Type, &c.Slug, &c.Title, &c.Summary, &c.Body, &c.CoverImage, &c.Status, &c.SortOrder, &c.PublishedAt, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read updated content")
			return
		}

		json.NewEncoder(w).Encode(c)
	}
}

func handleDeleteContent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			w.Header().Set("Content-Type", "application/json")
			writeError(w, http.StatusBadRequest, "missing id")
			return
		}

		result, err := db.Exec(`DELETE FROM content WHERE id = ?`, id)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			writeError(w, http.StatusInternalServerError, "failed to delete content")
			return
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			w.Header().Set("Content-Type", "application/json")
			writeError(w, http.StatusNotFound, "content not found")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func handlePublishContent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "missing id")
			return
		}

		result, err := db.Exec(
			`UPDATE content SET status = 'published', published_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			id,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to publish content")
			return
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			writeError(w, http.StatusNotFound, "content not found")
			return
		}

		var c Content
		err = db.QueryRow(
			`SELECT id, type, slug, title, summary, body, cover_image, status, sort_order, published_at, created_at, updated_at FROM content WHERE id = ?`,
			id,
		).Scan(&c.ID, &c.Type, &c.Slug, &c.Title, &c.Summary, &c.Body, &c.CoverImage, &c.Status, &c.SortOrder, &c.PublishedAt, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read content")
			return
		}

		json.NewEncoder(w).Encode(c)
	}
}

func handleDraftContent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "missing id")
			return
		}

		result, err := db.Exec(
			`UPDATE content SET status = 'draft', updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			id,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to draft content")
			return
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			writeError(w, http.StatusNotFound, "content not found")
			return
		}

		var c Content
		err = db.QueryRow(
			`SELECT id, type, slug, title, summary, body, cover_image, status, sort_order, published_at, created_at, updated_at FROM content WHERE id = ?`,
			id,
		).Scan(&c.ID, &c.Type, &c.Slug, &c.Title, &c.Summary, &c.Body, &c.CoverImage, &c.Status, &c.SortOrder, &c.PublishedAt, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read content")
			return
		}

		json.NewEncoder(w).Encode(c)
	}
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9-]`)
var multipleHyphens = regexp.MustCompile(`-{2,}`)

func slugify(title string) string {
	replacements := map[rune]rune{
		'ą': 'a', 'ć': 'c', 'ę': 'e', 'ł': 'l', 'ń': 'n',
		'ó': 'o', 'ś': 's', 'ź': 'z', 'ż': 'z',
		'Ą': 'a', 'Ć': 'c', 'Ę': 'e', 'Ł': 'l', 'Ń': 'n',
		'Ó': 'o', 'Ś': 's', 'Ź': 'z', 'Ż': 'z',
	}

	var b strings.Builder
	for _, r := range title {
		if replacement, ok := replacements[r]; ok {
			b.WriteRune(replacement)
		} else {
			b.WriteRune(r)
		}
	}

	result := strings.ToLower(b.String())
	result = strings.ReplaceAll(result, " ", "-")
	result = nonAlphanumeric.ReplaceAllString(result, "")
	result = multipleHyphens.ReplaceAllString(result, "-")
	result = strings.Trim(result, "-")

	// Fallback for empty slugs
	if result == "" {
		result = "untitled-" + strconv.FormatInt(0, 10)
	}

	return result
}
