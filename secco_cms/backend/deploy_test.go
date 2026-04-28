package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestBuildSnapshot_Empty(t *testing.T) {
	db := testDB(t)

	snap, err := buildSnapshot(db)
	if err != nil {
		t.Fatalf("buildSnapshot: %v", err)
	}

	if snap.ExportedAt == "" {
		t.Error("expected non-empty ExportedAt")
	}
	if len(snap.Articles) != 0 {
		t.Errorf("expected 0 articles, got %d", len(snap.Articles))
	}
	if len(snap.Projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(snap.Projects))
	}
	if snap.Process != nil {
		t.Error("expected nil process")
	}
}

func TestBuildSnapshot_OnlyPublishedContent(t *testing.T) {
	db := testDB(t)

	// Create articles: one published, one draft
	db.Exec(`INSERT INTO content (type, slug, title, summary, body, cover_image, status, published_at) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		"article", "published-article", "Published Article", "summary", "body", "cover.png", "published")
	db.Exec(`INSERT INTO content (type, slug, title, summary, body, cover_image, status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"article", "draft-article", "Draft Article", "draft summary", "draft body", "", "draft")

	// Create projects: one published, one draft
	db.Exec(`INSERT INTO content (type, slug, title, summary, body, cover_image, status, sort_order, published_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		"project", "published-project", "Published Project", "proj summary", "proj body", "proj.png", "published", 1)
	db.Exec(`INSERT INTO content (type, slug, title, summary, body, cover_image, status, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"project", "draft-project", "Draft Project", "", "", "", "draft", 2)

	// Create published process
	db.Exec(`INSERT INTO content (type, slug, title, summary, body, status, published_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		"process", "our-process", "Our Process", "", "process body", "published")

	snap, err := buildSnapshot(db)
	if err != nil {
		t.Fatalf("buildSnapshot: %v", err)
	}

	if len(snap.Articles) != 1 {
		t.Errorf("expected 1 published article, got %d", len(snap.Articles))
	}
	if snap.Articles[0].Slug != "published-article" {
		t.Errorf("expected slug 'published-article', got %q", snap.Articles[0].Slug)
	}
	if snap.Articles[0].Title != "Published Article" {
		t.Errorf("expected title 'Published Article', got %q", snap.Articles[0].Title)
	}

	if len(snap.Projects) != 1 {
		t.Errorf("expected 1 published project, got %d", len(snap.Projects))
	}
	if snap.Projects[0].Slug != "published-project" {
		t.Errorf("expected slug 'published-project', got %q", snap.Projects[0].Slug)
	}
	if snap.Projects[0].SortOrder != 1 {
		t.Errorf("expected sort_order 1, got %d", snap.Projects[0].SortOrder)
	}

	if snap.Process == nil {
		t.Fatal("expected process to be set")
	}
	if snap.Process.Title != "Our Process" {
		t.Errorf("expected process title 'Our Process', got %q", snap.Process.Title)
	}
	if snap.Process.Body != "process body" {
		t.Errorf("expected process body 'process body', got %q", snap.Process.Body)
	}
}

func TestBuildSnapshot_DraftProcessExcluded(t *testing.T) {
	db := testDB(t)

	db.Exec(`INSERT INTO content (type, slug, title, body, status) VALUES (?, ?, ?, ?, ?)`,
		"process", "draft-process", "Draft Process", "draft body", "draft")

	snap, err := buildSnapshot(db)
	if err != nil {
		t.Fatalf("buildSnapshot: %v", err)
	}

	if snap.Process != nil {
		t.Error("expected nil process for draft-only")
	}
}

func TestHandleDeploy(t *testing.T) {
	db := testDB(t)

	// Create some published content
	db.Exec(`INSERT INTO content (type, slug, title, summary, body, status, published_at) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		"article", "test-article", "Test Article", "summary", "body", "published")
	db.Exec(`INSERT INTO content (type, slug, title, summary, body, status, sort_order, published_at) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		"project", "test-project", "Test Project", "proj summary", "proj body", "published", 1)

	var uploadedKeys []string
	store := &mockObjectStore{
		putObjectFn: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			uploadedKeys = append(uploadedKeys, *params.Key)
			return &s3.PutObjectOutput{}, nil
		},
	}

	cfg := testImageConfig()
	cfg.githubRepo = "test/repo"

	req := httptest.NewRequest("POST", "/api/deploy", nil)
	rec := httptest.NewRecorder()

	handleDeploy(db, store, cfg)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify two uploads: timestamped + latest
	if len(uploadedKeys) != 2 {
		t.Fatalf("expected 2 uploads, got %d: %v", len(uploadedKeys), uploadedKeys)
	}

	hasTimestamped := false
	hasLatest := false
	for _, key := range uploadedKeys {
		if key == "snapshots/latest.json" {
			hasLatest = true
		} else if len(key) > len("snapshots/") {
			hasTimestamped = true
		}
	}
	if !hasTimestamped {
		t.Error("expected timestamped snapshot upload")
	}
	if !hasLatest {
		t.Error("expected latest.json upload")
	}

	// Parse response
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)

	if result["snapshot_url"] == nil || result["snapshot_url"] == "" {
		t.Error("expected snapshot_url in response")
	}
	if result["articles_count"] != float64(1) {
		t.Errorf("expected articles_count 1, got %v", result["articles_count"])
	}
	if result["projects_count"] != float64(1) {
		t.Errorf("expected projects_count 1, got %v", result["projects_count"])
	}
}

func TestHandleDeploy_UploadFailure(t *testing.T) {
	db := testDB(t)

	store := &mockObjectStore{
		putObjectFn: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return nil, fmt.Errorf("storage down")
		},
	}

	cfg := testImageConfig()

	req := httptest.NewRequest("POST", "/api/deploy", nil)
	rec := httptest.NewRecorder()

	handleDeploy(db, store, cfg)(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestHandleDeploy_RecordsDeploy(t *testing.T) {
	db := testDB(t)
	store := &mockObjectStore{}
	cfg := testImageConfig()
	cfg.githubRepo = "test/repo"

	req := httptest.NewRequest("POST", "/api/deploy", nil)
	rec := httptest.NewRecorder()

	handleDeploy(db, store, cfg)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify deploy record in DB
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM deploys`).Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 deploy record, got %d", count)
	}

	var status string
	db.QueryRow(`SELECT status FROM deploys ORDER BY id DESC LIMIT 1`).Scan(&status)
	if status != "triggered" {
		t.Errorf("expected status 'triggered', got %q", status)
	}
}

func TestHandleDeployStatus_NoPreviousDeploy(t *testing.T) {
	db := testDB(t)

	req := httptest.NewRequest("GET", "/api/deploy/status", nil)
	rec := httptest.NewRecorder()

	handleDeployStatus(db)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)

	if result["status"] != "ready" {
		t.Errorf("expected status 'ready', got %v", result["status"])
	}
	if result["last_deploy"] != nil {
		t.Errorf("expected nil last_deploy, got %v", result["last_deploy"])
	}
}

func TestHandleDeployStatus_WithPreviousDeploy(t *testing.T) {
	db := testDB(t)

	// Insert a deploy record
	db.Exec(`INSERT INTO deploys (status, snapshot_key, gh_run_url) VALUES (?, ?, ?)`,
		"triggered", "snapshots/2026-04-15T10-00-00Z.json", "https://github.com/test/repo/actions")

	// Insert some published content
	db.Exec(`INSERT INTO content (type, slug, title, body, status, published_at) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		"article", "test-art", "Test", "body", "published")

	req := httptest.NewRequest("GET", "/api/deploy/status", nil)
	rec := httptest.NewRecorder()

	handleDeployStatus(db)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)

	if result["status"] != "triggered" {
		t.Errorf("expected status 'triggered', got %v", result["status"])
	}
	if result["last_deploy"] == nil {
		t.Error("expected non-nil last_deploy")
	}
	if result["snapshot_key"] != "snapshots/2026-04-15T10-00-00Z.json" {
		t.Errorf("expected snapshot_key, got %v", result["snapshot_key"])
	}
	if result["published_articles"] != float64(1) {
		t.Errorf("expected published_articles 1, got %v", result["published_articles"])
	}
}

func TestHandleDeployStatus_HasChanges(t *testing.T) {
	db := testDB(t)

	// Insert old deploy with a fixed past timestamp
	db.Exec(`INSERT INTO deploys (status, snapshot_key, created_at) VALUES (?, ?, ?)`,
		"triggered", "snapshots/old.json", "2020-01-01 00:00:00")

	// Insert content with default updated_at (CURRENT_TIMESTAMP, which is now and thus after 2020)
	db.Exec(`INSERT INTO content (type, slug, title, body, status) VALUES (?, ?, ?, ?, ?)`,
		"article", "new-art", "New Article", "body", "draft")

	req := httptest.NewRequest("GET", "/api/deploy/status", nil)
	rec := httptest.NewRecorder()

	handleDeployStatus(db)(rec, req)

	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)

	if result["has_changes"] != true {
		t.Errorf("expected has_changes true, got %v", result["has_changes"])
	}
}
