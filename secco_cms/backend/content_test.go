package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"Projektowanie wnętrz", "projektowanie-wnetrz"},
		{"Łódzki projekt", "lodzki-projekt"},
		{"Ćwiczenie źródeł", "cwiczenie-zrodel"},
		{"  spaces  around  ", "spaces-around"},
		{"Special!@#chars$%^", "specialchars"},
		{"Żółta ścieżka", "zolta-sciezka"},
		{"Already-slugified", "already-slugified"},
		{"UPPERCASE TITLE", "uppercase-title"},
		{"multiple   spaces", "multiple-spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHandleCreateContent(t *testing.T) {
	db := testDB(t)

	input := ContentInput{
		Type:    "article",
		Title:   "Test Article",
		Summary: "A test summary",
		Body:    "Article body content",
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest("POST", "/api/content", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handleCreateContent(db)(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var c Content
	if err := json.NewDecoder(rec.Body).Decode(&c); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if c.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if c.Slug != "test-article" {
		t.Errorf("expected slug 'test-article', got %q", c.Slug)
	}
	if c.Status != "draft" {
		t.Errorf("expected status 'draft', got %q", c.Status)
	}
}

func TestHandleCreateContent_MissingTitle(t *testing.T) {
	db := testDB(t)

	input := ContentInput{Type: "article"}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest("POST", "/api/content", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handleCreateContent(db)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleCreateContent_InvalidType(t *testing.T) {
	db := testDB(t)

	input := ContentInput{Type: "invalid", Title: "Test"}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest("POST", "/api/content", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handleCreateContent(db)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleCreateContent_ProcessUniqueness(t *testing.T) {
	db := testDB(t)

	// Create first process
	input := ContentInput{Type: "process", Title: "Design Process"}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/api/content", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleCreateContent(db)(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("first process: expected 201, got %d", rec.Code)
	}

	// Try second process
	input2 := ContentInput{Type: "process", Title: "Another Process"}
	body2, _ := json.Marshal(input2)
	req2 := httptest.NewRequest("POST", "/api/content", bytes.NewReader(body2))
	rec2 := httptest.NewRecorder()
	handleCreateContent(db)(rec2, req2)

	if rec2.Code != http.StatusConflict {
		t.Errorf("second process: expected 409, got %d: %s", rec2.Code, rec2.Body.String())
	}
}

func TestHandleCreateContent_SlugUniqueness(t *testing.T) {
	db := testDB(t)

	input := ContentInput{Type: "article", Title: "Same Title"}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/api/content", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleCreateContent(db)(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("first create: expected 201, got %d", rec.Code)
	}

	// Same title => same slug => conflict
	body2, _ := json.Marshal(input)
	req2 := httptest.NewRequest("POST", "/api/content", bytes.NewReader(body2))
	rec2 := httptest.NewRecorder()
	handleCreateContent(db)(rec2, req2)

	if rec2.Code != http.StatusConflict {
		t.Errorf("duplicate slug: expected 409, got %d", rec2.Code)
	}
}

func createTestContent(t *testing.T, db *sql.DB, input ContentInput) Content {
	t.Helper()
	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/api/content", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleCreateContent(db)(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create failed: %d %s", rec.Code, rec.Body.String())
	}
	var c Content
	json.NewDecoder(rec.Body).Decode(&c)
	return c
}

func TestHandleGetContent(t *testing.T) {
	db := testDB(t)

	created := createTestContent(t, db, ContentInput{
		Type:  "article",
		Title: "Get Test",
		Body:  "Full body here",
	})

	req := httptest.NewRequest("GET", "/api/content/"+json_id(created.ID), nil)
	req.SetPathValue("id", json_id(created.ID))
	rec := httptest.NewRecorder()

	handleGetContent(db)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var c Content
	json.NewDecoder(rec.Body).Decode(&c)

	if c.Body != "Full body here" {
		t.Errorf("expected body 'Full body here', got %q", c.Body)
	}
}

func TestHandleGetContent_NotFound(t *testing.T) {
	db := testDB(t)

	req := httptest.NewRequest("GET", "/api/content/999", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()

	handleGetContent(db)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandleListContent(t *testing.T) {
	db := testDB(t)

	createTestContent(t, db, ContentInput{Type: "article", Title: "Article 1"})
	createTestContent(t, db, ContentInput{Type: "project", Title: "Project 1"})
	createTestContent(t, db, ContentInput{Type: "article", Title: "Article 2"})

	// List all
	req := httptest.NewRequest("GET", "/api/content", nil)
	rec := httptest.NewRecorder()
	handleListContent(db)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var items []Content
	json.NewDecoder(rec.Body).Decode(&items)
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}

	// Body should not be included in list
	for _, item := range items {
		if item.Body != "" {
			t.Errorf("list should not include body, got %q", item.Body)
		}
	}

	// Filter by type
	req2 := httptest.NewRequest("GET", "/api/content?type=article", nil)
	rec2 := httptest.NewRecorder()
	handleListContent(db)(rec2, req2)

	var articles []Content
	json.NewDecoder(rec2.Body).Decode(&articles)
	if len(articles) != 2 {
		t.Errorf("expected 2 articles, got %d", len(articles))
	}
}

func TestHandleUpdateContent(t *testing.T) {
	db := testDB(t)

	created := createTestContent(t, db, ContentInput{
		Type:  "article",
		Title: "Original Title",
		Body:  "Original body",
	})

	updated := ContentInput{
		Type:    "article",
		Title:   "Updated Title",
		Slug:    "updated-title",
		Body:    "Updated body",
		Summary: "Now with summary",
	}
	body, _ := json.Marshal(updated)
	req := httptest.NewRequest("PUT", "/api/content/"+json_id(created.ID), bytes.NewReader(body))
	req.SetPathValue("id", json_id(created.ID))
	rec := httptest.NewRecorder()

	handleUpdateContent(db)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var c Content
	json.NewDecoder(rec.Body).Decode(&c)

	if c.Title != "Updated Title" {
		t.Errorf("title not updated, got %q", c.Title)
	}
	if c.Body != "Updated body" {
		t.Errorf("body not updated, got %q", c.Body)
	}
}

func TestHandleUpdateContent_NotFound(t *testing.T) {
	db := testDB(t)

	input := ContentInput{Type: "article", Title: "Test"}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest("PUT", "/api/content/999", bytes.NewReader(body))
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()

	handleUpdateContent(db)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandleDeleteContent(t *testing.T) {
	db := testDB(t)

	created := createTestContent(t, db, ContentInput{Type: "article", Title: "To Delete"})

	req := httptest.NewRequest("DELETE", "/api/content/"+json_id(created.ID), nil)
	req.SetPathValue("id", json_id(created.ID))
	rec := httptest.NewRecorder()

	handleDeleteContent(db)(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}

	// Verify deleted
	req2 := httptest.NewRequest("GET", "/api/content/"+json_id(created.ID), nil)
	req2.SetPathValue("id", json_id(created.ID))
	rec2 := httptest.NewRecorder()
	handleGetContent(db)(rec2, req2)

	if rec2.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", rec2.Code)
	}
}

func TestHandleDeleteContent_NotFound(t *testing.T) {
	db := testDB(t)

	req := httptest.NewRequest("DELETE", "/api/content/999", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()

	handleDeleteContent(db)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandlePublishContent(t *testing.T) {
	db := testDB(t)

	created := createTestContent(t, db, ContentInput{Type: "article", Title: "To Publish"})

	req := httptest.NewRequest("PUT", "/api/content/"+json_id(created.ID)+"/publish", nil)
	req.SetPathValue("id", json_id(created.ID))
	rec := httptest.NewRecorder()

	handlePublishContent(db)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var c Content
	json.NewDecoder(rec.Body).Decode(&c)

	if c.Status != "published" {
		t.Errorf("expected status 'published', got %q", c.Status)
	}
	if c.PublishedAt == nil {
		t.Error("expected published_at to be set")
	}
}

func TestHandleDraftContent(t *testing.T) {
	db := testDB(t)

	created := createTestContent(t, db, ContentInput{Type: "article", Title: "To Draft"})

	// Publish first
	req := httptest.NewRequest("PUT", "/api/content/"+json_id(created.ID)+"/publish", nil)
	req.SetPathValue("id", json_id(created.ID))
	rec := httptest.NewRecorder()
	handlePublishContent(db)(rec, req)

	// Then draft
	req2 := httptest.NewRequest("PUT", "/api/content/"+json_id(created.ID)+"/draft", nil)
	req2.SetPathValue("id", json_id(created.ID))
	rec2 := httptest.NewRecorder()
	handleDraftContent(db)(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}

	var c Content
	json.NewDecoder(rec2.Body).Decode(&c)

	if c.Status != "draft" {
		t.Errorf("expected status 'draft', got %q", c.Status)
	}
}

func TestHandleListContent_FilterByStatus(t *testing.T) {
	db := testDB(t)

	c1 := createTestContent(t, db, ContentInput{Type: "article", Title: "Draft One"})
	createTestContent(t, db, ContentInput{Type: "article", Title: "Draft Two"})

	// Publish one
	req := httptest.NewRequest("PUT", "/api/content/"+json_id(c1.ID)+"/publish", nil)
	req.SetPathValue("id", json_id(c1.ID))
	rec := httptest.NewRecorder()
	handlePublishContent(db)(rec, req)

	// Filter published
	req2 := httptest.NewRequest("GET", "/api/content?status=published", nil)
	rec2 := httptest.NewRecorder()
	handleListContent(db)(rec2, req2)

	var items []Content
	json.NewDecoder(rec2.Body).Decode(&items)
	if len(items) != 1 {
		t.Errorf("expected 1 published, got %d", len(items))
	}

	// Filter draft
	req3 := httptest.NewRequest("GET", "/api/content?status=draft", nil)
	rec3 := httptest.NewRecorder()
	handleListContent(db)(rec3, req3)

	var drafts []Content
	json.NewDecoder(rec3.Body).Decode(&drafts)
	if len(drafts) != 1 {
		t.Errorf("expected 1 draft, got %d", len(drafts))
	}
}

func json_id(id int64) string {
	return strconv.FormatInt(id, 10)
}
