package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed all:build
var staticFiles embed.FS

func main() {
	port := envOrDefault("PORT", "8081")

	dbPath := envOrDefault("DATABASE_PATH", "./data/cms.db")
	db, err := InitDB(dbPath)
	if err != nil {
		log.Fatal("failed to init database:", err)
	}
	defer db.Close()

	cfg := loadConfig()

	s3Client := newR2Client(cfg)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /api/health", handleHealth())

	// Auth routes
	mux.HandleFunc("POST /api/auth/login", handleLogin(cfg))
	mux.HandleFunc("POST /api/auth/logout", handleLogout())
	mux.HandleFunc("GET /api/auth/check", authMiddleware(cfg, handleAuthCheck()))

	// Content CRUD (all protected)
	mux.HandleFunc("GET /api/content", authMiddleware(cfg, handleListContent(db)))
	mux.HandleFunc("GET /api/content/{id}", authMiddleware(cfg, handleGetContent(db)))
	mux.HandleFunc("POST /api/content", authMiddleware(cfg, handleCreateContent(db)))
	mux.HandleFunc("PUT /api/content/{id}", authMiddleware(cfg, handleUpdateContent(db)))
	mux.HandleFunc("DELETE /api/content/{id}", authMiddleware(cfg, handleDeleteContent(db)))
	mux.HandleFunc("PUT /api/content/{id}/publish", authMiddleware(cfg, handlePublishContent(db)))
	mux.HandleFunc("PUT /api/content/{id}/draft", authMiddleware(cfg, handleDraftContent(db)))

	// Image management (all protected)
	mux.HandleFunc("POST /api/images/upload", authMiddleware(cfg, handleUploadImage(db, s3Client, cfg)))
	mux.HandleFunc("GET /api/images", authMiddleware(cfg, handleListImages(db)))
	mux.HandleFunc("PUT /api/images/{id}", authMiddleware(cfg, handleUpdateImageAlt(db)))
	mux.HandleFunc("DELETE /api/images/{id}", authMiddleware(cfg, handleDeleteImage(db, s3Client, cfg)))

	// Deploy (protected)
	mux.HandleFunc("POST /api/deploy", authMiddleware(cfg, handleDeploy(db, s3Client, cfg)))
	mux.HandleFunc("GET /api/deploy/status", authMiddleware(cfg, handleDeployStatus(db)))

	// Static files (SvelteKit build)
	staticFS, err := fs.Sub(staticFiles, "build")
	if err != nil {
		log.Fatal("failed to create sub filesystem:", err)
	}
	fileServer := http.FileServer(http.FS(staticFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path != "/" {
			cleanPath := strings.TrimPrefix(path, "/")
			if f, err := staticFS.(fs.ReadFileFS).ReadFile(cleanPath); err == nil {
				_ = f
				fileServer.ServeHTTP(w, r)
				return
			}
		} else {
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback
		r.URL.Path = "/index.html"
		fileServer.ServeHTTP(w, r)
	})

	log.Printf("secco-cms listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

type config struct {
	adminUser        string
	adminPassHash    string
	jwtSecret        string
	dbPath           string
	r2AccountID      string
	r2AccessKeyID    string
	r2SecretAccessKey string
	r2BucketName     string
	r2PublicURL      string
	githubToken      string
	githubRepo       string
	githubWorkflowID string
}

func loadConfig() config {
	return config{
		adminUser:        envOrDefault("CMS_ADMIN_USER", "admin"),
		adminPassHash:    os.Getenv("CMS_ADMIN_PASS_HASH"),
		jwtSecret:        os.Getenv("CMS_JWT_SECRET"),
		dbPath:           envOrDefault("DATABASE_PATH", "./data/cms.db"),
		r2AccountID:      os.Getenv("R2_ACCOUNT_ID"),
		r2AccessKeyID:    os.Getenv("R2_ACCESS_KEY_ID"),
		r2SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		r2BucketName:     envOrDefault("R2_BUCKET_NAME", "secco-assets"),
		r2PublicURL:      envOrDefault("R2_PUBLIC_URL", "https://assets.seccostudio.com"),
		githubToken:      os.Getenv("GITHUB_TOKEN"),
		githubRepo:       envOrDefault("GITHUB_REPO", "morawskiOZ/secco-monorepo"),
		githubWorkflowID: envOrDefault("GITHUB_WORKFLOW_ID", "build-lp.yml"),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}
}
