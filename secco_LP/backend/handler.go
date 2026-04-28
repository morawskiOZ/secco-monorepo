package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const maxUploadSize = 100 << 20 // 100MB

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{requests: make(map[string][]time.Time)}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Hour)

	// Clean old entries
	times := rl.requests[ip]
	valid := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= 5 {
		rl.requests[ip] = valid
		return false
	}

	rl.requests[ip] = append(valid, now)
	return true
}

func handleSubmit(cfg config, limiter *rateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ip := clientIP(r)

		if !limiter.allow(ip) {
			log.Printf("rate limited: %s", ip)
			writeError(w, http.StatusTooManyRequests, "Zbyt wiele prób. Spróbuj ponownie za godzinę.")
			return
		}

		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			writeError(w, http.StatusBadRequest, "Nieprawidłowe dane formularza.")
			return
		}

		// Verify Turnstile token
		token := r.FormValue("cf-turnstile-response")
		if ok, err := verifyTurnstile(cfg.turnstileKey, token, ip); err != nil || !ok {
			log.Printf("turnstile verification failed: ip=%s err=%v", ip, err)
			writeError(w, http.StatusForbidden, "Weryfikacja nie powiodła się. Odśwież stronę i spróbuj ponownie.")
			return
		}

		// Validate required fields
		firstName := strings.TrimSpace(r.FormValue("first_name"))
		email := strings.TrimSpace(r.FormValue("email"))
		if firstName == "" || email == "" {
			writeError(w, http.StatusBadRequest, "Imię i email są wymagane.")
			return
		}
		if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
			writeError(w, http.StatusBadRequest, "Nieprawidłowy adres email.")
			return
		}

		// Collect all form answers
		answers := make(map[string]string)
		for key, values := range r.MultipartForm.Value {
			if key == "cf-turnstile-response" {
				continue
			}
			answers[key] = strings.Join(values, ", ")
		}

		// Collect file attachments
		var attachments []attachment
		for _, fileHeaders := range r.MultipartForm.File {
			for _, fh := range fileHeaders {
				f, err := fh.Open()
				if err != nil {
					continue
				}
				data, err := io.ReadAll(f)
				f.Close()
				if err != nil {
					continue
				}
				attachments = append(attachments, attachment{
					Name:        fh.Filename,
					ContentType: fh.Header.Get("Content-Type"),
					Data:        data,
				})
			}
		}

		lastName := strings.TrimSpace(r.FormValue("last_name"))
		sub := submission{
			FirstName:   firstName,
			LastName:    lastName,
			Email:       email,
			Answers:     answers,
			Attachments: attachments,
		}

		if err := sendEmail(cfg, sub); err != nil {
			log.Printf("email send failed: %v", err)
			writeError(w, http.StatusInternalServerError, "Nie udało się wysłać wiadomości. Spróbuj ponownie później.")
			return
		}

		log.Printf("submission received: name=%s email=%s attachments=%d", firstName, email, len(attachments))
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}
}

func handlePreferencesSubmit(cfg config, limiter *rateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ip := clientIP(r)

		if !limiter.allow(ip) {
			log.Printf("rate limited: %s", ip)
			writeError(w, http.StatusTooManyRequests, "Zbyt wiele prób. Spróbuj ponownie za godzinę.")
			return
		}

		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			writeError(w, http.StatusBadRequest, "Nieprawidłowe dane formularza.")
			return
		}

		token := r.FormValue("cf-turnstile-response")
		if ok, err := verifyTurnstile(cfg.turnstileKey, token, ip); err != nil || !ok {
			log.Printf("turnstile verification failed: ip=%s err=%v", ip, err)
			writeError(w, http.StatusForbidden, "Weryfikacja nie powiodła się. Odśwież stronę i spróbuj ponownie.")
			return
		}

		firstName := strings.TrimSpace(r.FormValue("first_name"))
		email := strings.TrimSpace(r.FormValue("email"))
		if firstName == "" || email == "" {
			writeError(w, http.StatusBadRequest, "Imię i email są wymagane.")
			return
		}
		if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
			writeError(w, http.StatusBadRequest, "Nieprawidłowy adres email.")
			return
		}

		answers := make(map[string]string)
		for key, values := range r.MultipartForm.Value {
			if key == "cf-turnstile-response" {
				continue
			}
			answers[key] = strings.Join(values, ", ")
		}

		var attachments []attachment
		for _, fileHeaders := range r.MultipartForm.File {
			for _, fh := range fileHeaders {
				f, err := fh.Open()
				if err != nil {
					continue
				}
				data, err := io.ReadAll(f)
				f.Close()
				if err != nil {
					continue
				}
				attachments = append(attachments, attachment{
					Name:        fh.Filename,
					ContentType: fh.Header.Get("Content-Type"),
					Data:        data,
				})
			}
		}

		lastName := strings.TrimSpace(r.FormValue("last_name"))
		sub := submission{
			FirstName:   firstName,
			LastName:    lastName,
			Email:       email,
			Answers:     answers,
			Attachments: attachments,
		}

		if err := sendPreferencesEmail(cfg, sub); err != nil {
			log.Printf("preferences email send failed: %v", err)
			writeError(w, http.StatusInternalServerError, "Nie udało się wysłać wiadomości. Spróbuj ponownie później.")
			return
		}

		log.Printf("preferences submission received: name=%s email=%s attachments=%d", firstName, email, len(attachments))
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
