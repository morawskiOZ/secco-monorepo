package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Snapshot struct {
	ExportedAt string            `json:"exported_at"`
	Articles   []SnapshotArticle `json:"articles"`
	Projects   []SnapshotProject `json:"projects"`
	Process    *SnapshotProcess  `json:"process"`
}

type SnapshotArticle struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Body        string `json:"body"`
	CoverImage  string `json:"cover_image"`
	Tags        string `json:"tags"`
	PublishedAt string `json:"published_at"`
}

type SnapshotProject struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Body        string `json:"body"`
	CoverImage  string `json:"cover_image"`
	SortOrder   int    `json:"sort_order"`
	PublishedAt string `json:"published_at"`
}

type SnapshotProcess struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func buildSnapshot(db *sql.DB) (Snapshot, error) {
	snap := Snapshot{
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Articles:   []SnapshotArticle{},
		Projects:   []SnapshotProject{},
	}

	// Articles
	rows, err := db.Query(`SELECT slug, title, summary, body, cover_image, published_at FROM content WHERE type = 'article' AND status = 'published' ORDER BY published_at DESC`)
	if err != nil {
		return snap, fmt.Errorf("query articles: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var a SnapshotArticle
		var publishedAt *string
		if err := rows.Scan(&a.Slug, &a.Title, &a.Summary, &a.Body, &a.CoverImage, &publishedAt); err != nil {
			return snap, fmt.Errorf("scan article: %w", err)
		}
		if publishedAt != nil {
			a.PublishedAt = *publishedAt
		}
		snap.Articles = append(snap.Articles, a)
	}

	// Projects
	projRows, err := db.Query(`SELECT slug, title, summary, body, cover_image, sort_order, published_at FROM content WHERE type = 'project' AND status = 'published' ORDER BY sort_order, published_at DESC`)
	if err != nil {
		return snap, fmt.Errorf("query projects: %w", err)
	}
	defer projRows.Close()

	for projRows.Next() {
		var p SnapshotProject
		var publishedAt *string
		if err := projRows.Scan(&p.Slug, &p.Title, &p.Summary, &p.Body, &p.CoverImage, &p.SortOrder, &publishedAt); err != nil {
			return snap, fmt.Errorf("scan project: %w", err)
		}
		if publishedAt != nil {
			p.PublishedAt = *publishedAt
		}
		snap.Projects = append(snap.Projects, p)
	}

	// Process (at most one)
	var proc SnapshotProcess
	err = db.QueryRow(`SELECT title, body FROM content WHERE type = 'process' AND status = 'published' LIMIT 1`).Scan(&proc.Title, &proc.Body)
	if err == nil {
		snap.Process = &proc
	}

	return snap, nil
}

func handleDeploy(db *sql.DB, store objectStore, cfg config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		snap, err := buildSnapshot(db)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to build snapshot: %v", err))
			return
		}

		data, err := json.MarshalIndent(snap, "", "  ")
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to marshal snapshot")
			return
		}

		timestamp := time.Now().UTC().Format("2006-01-02T15-04-05Z")
		timestampKey := fmt.Sprintf("snapshots/%s.json", timestamp)
		latestKey := "snapshots/latest.json"

		// Upload timestamped backup
		_, err = store.PutObject(r.Context(), &s3.PutObjectInput{
			Bucket:      aws.String(cfg.r2BucketName),
			Key:         aws.String(timestampKey),
			Body:        bytes.NewReader(data),
			ContentType: aws.String("application/json"),
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to upload snapshot backup")
			return
		}

		// Upload latest
		_, err = store.PutObject(r.Context(), &s3.PutObjectInput{
			Bucket:      aws.String(cfg.r2BucketName),
			Key:         aws.String(latestKey),
			Body:        bytes.NewReader(data),
			ContentType: aws.String("application/json"),
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to upload latest snapshot")
			return
		}

		// Trigger GitHub Actions
		ghRunURL := ""
		if cfg.githubToken != "" && cfg.githubRepo != "" && cfg.githubWorkflowID != "" {
			ghRunURL, err = triggerGitHubWorkflow(cfg, timestamp)
			if err != nil {
				writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to trigger deploy: %v", err))
				return
			}
		}

		// Record deploy in DB
		db.Exec(
			`INSERT INTO deploys (status, snapshot_key, gh_run_url) VALUES (?, ?, ?)`,
			"triggered", timestampKey, ghRunURL,
		)

		snapshotURL := cfg.r2PublicURL + "/" + timestampKey
		if ghRunURL == "" {
			ghRunURL = fmt.Sprintf("https://github.com/%s/actions", cfg.githubRepo)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"snapshot_url":    snapshotURL,
			"run_url":         ghRunURL,
			"articles_count":  len(snap.Articles),
			"projects_count":  len(snap.Projects),
		})
	}
}

func triggerGitHubWorkflow(cfg config, timestamp string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/actions/workflows/%s/dispatches", cfg.githubRepo, cfg.githubWorkflowID)

	version := fmt.Sprintf("0.0.%d", time.Now().Unix())

	payload := map[string]interface{}{
		"ref": "main",
		"inputs": map[string]string{
			"version":            version,
			"snapshot_timestamp": timestamp,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.githubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("dispatch request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return "", fmt.Errorf("github returned status %d", resp.StatusCode)
	}

	runURL := fmt.Sprintf("https://github.com/%s/actions", cfg.githubRepo)
	return runURL, nil
}

func handleDeployStatus(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var lastDeploy *string
		var status string
		var snapshotKey, ghRunURL sql.NullString

		err := db.QueryRow(
			`SELECT status, snapshot_key, gh_run_url, created_at FROM deploys ORDER BY created_at DESC LIMIT 1`,
		).Scan(&status, &snapshotKey, &ghRunURL, &lastDeploy)

		if err == sql.ErrNoRows {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":      "ready",
				"last_deploy": nil,
			})
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to query deploy status")
			return
		}

		result := map[string]interface{}{
			"status":      status,
			"last_deploy": lastDeploy,
		}
		if snapshotKey.Valid {
			result["snapshot_key"] = snapshotKey.String
		}
		if ghRunURL.Valid {
			result["run_url"] = ghRunURL.String
		}

		// Count published content for context
		var articleCount, projectCount int
		db.QueryRow(`SELECT COUNT(*) FROM content WHERE type = 'article' AND status = 'published'`).Scan(&articleCount)
		db.QueryRow(`SELECT COUNT(*) FROM content WHERE type = 'project' AND status = 'published'`).Scan(&projectCount)

		result["published_articles"] = articleCount
		result["published_projects"] = projectCount

		// Check if there are changes since last deploy
		if lastDeploy != nil {
			var changedCount int
			db.QueryRow(`SELECT COUNT(*) FROM content WHERE updated_at > ?`, *lastDeploy).Scan(&changedCount)
			result["has_changes"] = changedCount > 0
		} else {
			result["has_changes"] = true
		}

		json.NewEncoder(w).Encode(result)
	}
}

