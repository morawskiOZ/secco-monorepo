package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed all:build
var staticFiles embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cfg := loadConfig()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	limiter := newRateLimiter()
	mux.HandleFunc("POST /api/submit", handleSubmit(cfg, limiter))
	mux.HandleFunc("POST /api/submit-preferences", handlePreferencesSubmit(cfg, limiter))

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

		// SPA fallback: serve 200.html for client-side routes
		r.URL.Path = "/200.html"
		fileServer.ServeHTTP(w, r)
	})

	log.Printf("secco-studio listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

type config struct {
	smtpHost       string
	smtpPort       string
	smtpUser       string
	smtpPass       string
	recipientEmail string
	turnstileKey   string
}

func loadConfig() config {
	return config{
		smtpHost:       envOrDefault("SMTP_HOST", "smtp.gmail.com"),
		smtpPort:       envOrDefault("SMTP_PORT", "587"),
		smtpUser:       os.Getenv("SMTP_USER"),
		smtpPass:       os.Getenv("SMTP_PASS"),
		recipientEmail: os.Getenv("RECIPIENT_EMAIL"),
		turnstileKey:   os.Getenv("TURNSTILE_SECRET_KEY"),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
