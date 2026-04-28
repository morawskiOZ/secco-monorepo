package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
)

// objectStore abstracts S3/R2 operations for testing.
type objectStore interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

type Image struct {
	ID          int64  `json:"id"`
	R2Key       string `json:"r2_key"`
	Filename    string `json:"filename"`
	AltText     string `json:"alt_text"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	SizeBytes   int64  `json:"size_bytes"`
	ContentType string `json:"content_type"`
	PublicURL   string `json:"public_url"`
	CreatedAt   string `json:"created_at"`
}

var validImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

var unsafeChars = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = unsafeChars.ReplaceAllString(name, "_")
	if name == "" || name == "." {
		name = "upload"
	}
	return name
}

func newR2Client(cfg config) *s3.Client {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.r2AccountID),
		}, nil
	})

	awsCfg, _ := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithEndpointResolverWithOptions(r2Resolver),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.r2AccessKeyID, cfg.r2SecretAccessKey, "",
		)),
		awsconfig.WithRegion("auto"),
	)

	return s3.NewFromConfig(awsCfg)
}

func handleUploadImage(db *sql.DB, store objectStore, cfg config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := r.ParseMultipartForm(20 << 20); err != nil {
			writeError(w, http.StatusBadRequest, "file too large or invalid form")
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			writeError(w, http.StatusBadRequest, "missing file field")
			return
		}
		defer file.Close()

		contentType := header.Header.Get("Content-Type")
		if !validImageTypes[contentType] {
			writeError(w, http.StatusBadRequest, "unsupported file type: must be jpeg, png, or webp")
			return
		}

		// Read file into memory for dimension detection and upload
		data, err := io.ReadAll(file)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read file")
			return
		}

		// Decode image dimensions
		imgCfg, _, err := image.DecodeConfig(bytes.NewReader(data))
		var width, height int
		if err == nil {
			width = imgCfg.Width
			height = imgCfg.Height
		}

		sanitized := sanitizeFilename(header.Filename)
		r2Key := fmt.Sprintf("images/%s-%s", uuid.New().String(), sanitized)

		// Upload to R2
		_, err = store.PutObject(r.Context(), &s3.PutObjectInput{
			Bucket:      aws.String(cfg.r2BucketName),
			Key:         aws.String(r2Key),
			Body:        bytes.NewReader(data),
			ContentType: aws.String(contentType),
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to upload to storage")
			return
		}

		publicURL := cfg.r2PublicURL + "/" + r2Key
		altText := r.FormValue("alt_text")

		result, err := db.Exec(
			`INSERT INTO images (r2_key, filename, alt_text, width, height, size_bytes, content_type, public_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			r2Key, header.Filename, altText, width, height, int64(len(data)), contentType, publicURL,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save image record")
			return
		}

		id, _ := result.LastInsertId()

		var img Image
		err = db.QueryRow(
			`SELECT id, r2_key, filename, alt_text, width, height, size_bytes, content_type, public_url, created_at FROM images WHERE id = ?`,
			id,
		).Scan(&img.ID, &img.R2Key, &img.Filename, &img.AltText, &img.Width, &img.Height, &img.SizeBytes, &img.ContentType, &img.PublicURL, &img.CreatedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read created image")
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(img)
	}
}

func handleListImages(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
		if perPage < 1 || perPage > 100 {
			perPage = 50
		}
		offset := (page - 1) * perPage

		query := `SELECT id, r2_key, filename, alt_text, width, height, size_bytes, content_type, public_url, created_at FROM images`
		var args []interface{}

		if search := r.URL.Query().Get("search"); search != "" {
			query += " WHERE filename LIKE ?"
			args = append(args, "%"+search+"%")
		}

		query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
		args = append(args, perPage, offset)

		rows, err := db.Query(query, args...)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to query images")
			return
		}
		defer rows.Close()

		images := []Image{}
		for rows.Next() {
			var img Image
			if err := rows.Scan(&img.ID, &img.R2Key, &img.Filename, &img.AltText, &img.Width, &img.Height, &img.SizeBytes, &img.ContentType, &img.PublicURL, &img.CreatedAt); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to scan image")
				return
			}
			images = append(images, img)
		}

		json.NewEncoder(w).Encode(images)
	}
}

func handleDeleteImage(db *sql.DB, store objectStore, cfg config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			w.Header().Set("Content-Type", "application/json")
			writeError(w, http.StatusBadRequest, "missing id")
			return
		}

		var img Image
		err := db.QueryRow(`SELECT id, r2_key FROM images WHERE id = ?`, id).Scan(&img.ID, &img.R2Key)
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			writeError(w, http.StatusNotFound, "image not found")
			return
		}
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			writeError(w, http.StatusInternalServerError, "failed to query image")
			return
		}

		// Delete from R2
		_, err = store.DeleteObject(r.Context(), &s3.DeleteObjectInput{
			Bucket: aws.String(cfg.r2BucketName),
			Key:    aws.String(img.R2Key),
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			writeError(w, http.StatusInternalServerError, "failed to delete from storage")
			return
		}

		// Delete from DB (images + content_images junction)
		db.Exec(`DELETE FROM content_images WHERE image_id = ?`, img.ID)
		_, err = db.Exec(`DELETE FROM images WHERE id = ?`, img.ID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			writeError(w, http.StatusInternalServerError, "failed to delete image record")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func handleUpdateImageAlt(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		id := r.PathValue("id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "missing id")
			return
		}

		var input struct {
			AltText string `json:"alt_text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		result, err := db.Exec(`UPDATE images SET alt_text = ? WHERE id = ?`, input.AltText, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update image")
			return
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			writeError(w, http.StatusNotFound, "image not found")
			return
		}

		var img Image
		err = db.QueryRow(
			`SELECT id, r2_key, filename, alt_text, width, height, size_bytes, content_type, public_url, created_at FROM images WHERE id = ?`,
			id,
		).Scan(&img.ID, &img.R2Key, &img.Filename, &img.AltText, &img.Width, &img.Height, &img.SizeBytes, &img.ContentType, &img.PublicURL, &img.CreatedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read updated image")
			return
		}

		json.NewEncoder(w).Encode(img)
	}
}
