package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// mockObjectStore implements objectStore for testing.
type mockObjectStore struct {
	putObjectFn    func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	deleteObjectFn func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	headObjectFn   func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}

func (m *mockObjectStore) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if m.putObjectFn != nil {
		return m.putObjectFn(ctx, params, optFns...)
	}
	return &s3.PutObjectOutput{}, nil
}

func (m *mockObjectStore) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	if m.deleteObjectFn != nil {
		return m.deleteObjectFn(ctx, params, optFns...)
	}
	return &s3.DeleteObjectOutput{}, nil
}

func (m *mockObjectStore) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	if m.headObjectFn != nil {
		return m.headObjectFn(ctx, params, optFns...)
	}
	return &s3.HeadObjectOutput{}, nil
}

func testImageConfig() config {
	return config{
		r2BucketName: "test-bucket",
		r2PublicURL:  "https://test.example.com",
	}
}

// createPNGImage creates a minimal valid PNG image in memory.
func createPNGImage(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

// createMultipartRequest builds a multipart upload request with the given file data.
func createMultipartRequest(t *testing.T, filename, contentType string, data []byte, extraFields map[string]string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
		t.Fatalf("copy file data: %v", err)
	}

	for k, v := range extraFields {
		writer.WriteField(k, v)
	}
	writer.Close()

	req := httptest.NewRequest("POST", "/api/images/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Override the file's content type in the multipart header
	// The multipart writer sets application/octet-stream by default.
	// We need to manipulate the raw multipart to set the correct content type.
	// Easier approach: rebuild with correct content type header.
	body.Reset()
	writer = multipart.NewWriter(&body)

	h := make(map[string][]string)
	h["Content-Disposition"] = []string{fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename)}
	h["Content-Type"] = []string{contentType}
	part, err = writer.CreatePart(h)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
		t.Fatalf("copy file data: %v", err)
	}

	for k, v := range extraFields {
		writer.WriteField(k, v)
	}
	writer.Close()

	req = httptest.NewRequest("POST", "/api/images/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestHandleUploadImage(t *testing.T) {
	db := testDB(t)
	store := &mockObjectStore{}
	cfg := testImageConfig()

	imgData := createPNGImage(t, 100, 50)
	req := createMultipartRequest(t, "test.png", "image/png", imgData, map[string]string{
		"alt_text": "A test image",
	})
	rec := httptest.NewRecorder()

	handleUploadImage(db, store, cfg)(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var img Image
	if err := json.NewDecoder(rec.Body).Decode(&img); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if img.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if img.Filename != "test.png" {
		t.Errorf("expected filename 'test.png', got %q", img.Filename)
	}
	if img.AltText != "A test image" {
		t.Errorf("expected alt_text 'A test image', got %q", img.AltText)
	}
	if img.Width != 100 {
		t.Errorf("expected width 100, got %d", img.Width)
	}
	if img.Height != 50 {
		t.Errorf("expected height 50, got %d", img.Height)
	}
	if img.ContentType != "image/png" {
		t.Errorf("expected content type 'image/png', got %q", img.ContentType)
	}
	if img.SizeBytes == 0 {
		t.Error("expected non-zero size_bytes")
	}
	if img.PublicURL == "" {
		t.Error("expected non-empty public_url")
	}
}

func TestHandleUploadImage_InvalidContentType(t *testing.T) {
	db := testDB(t)
	store := &mockObjectStore{}
	cfg := testImageConfig()

	req := createMultipartRequest(t, "test.gif", "image/gif", []byte("fake gif data"), nil)
	rec := httptest.NewRecorder()

	handleUploadImage(db, store, cfg)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleUploadImage_MissingFile(t *testing.T) {
	db := testDB(t)
	store := &mockObjectStore{}
	cfg := testImageConfig()

	req := httptest.NewRequest("POST", "/api/images/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	rec := httptest.NewRecorder()

	handleUploadImage(db, store, cfg)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleUploadImage_StorageFailure(t *testing.T) {
	db := testDB(t)
	store := &mockObjectStore{
		putObjectFn: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return nil, fmt.Errorf("storage unavailable")
		},
	}
	cfg := testImageConfig()

	imgData := createPNGImage(t, 10, 10)
	req := createMultipartRequest(t, "test.png", "image/png", imgData, nil)
	rec := httptest.NewRecorder()

	handleUploadImage(db, store, cfg)(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleUploadImage_VerifiesR2Upload(t *testing.T) {
	db := testDB(t)
	var uploadedKey string
	var uploadedBucket string
	store := &mockObjectStore{
		putObjectFn: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			uploadedKey = *params.Key
			uploadedBucket = *params.Bucket
			return &s3.PutObjectOutput{}, nil
		},
	}
	cfg := testImageConfig()

	imgData := createPNGImage(t, 10, 10)
	req := createMultipartRequest(t, "my-photo.png", "image/png", imgData, nil)
	rec := httptest.NewRecorder()

	handleUploadImage(db, store, cfg)(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	if uploadedBucket != "test-bucket" {
		t.Errorf("expected bucket 'test-bucket', got %q", uploadedBucket)
	}
	if uploadedKey == "" {
		t.Error("expected non-empty upload key")
	}
}


func TestHandleListImages(t *testing.T) {
	db := testDB(t)

	// Insert test images directly
	db.Exec(`INSERT INTO images (r2_key, filename, alt_text, width, height, size_bytes, content_type, public_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"images/a.png", "alpha.png", "", 100, 100, 1024, "image/png", "https://test.example.com/images/a.png")
	db.Exec(`INSERT INTO images (r2_key, filename, alt_text, width, height, size_bytes, content_type, public_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"images/b.png", "beta.png", "", 200, 200, 2048, "image/png", "https://test.example.com/images/b.png")
	db.Exec(`INSERT INTO images (r2_key, filename, alt_text, width, height, size_bytes, content_type, public_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"images/c.png", "charlie.png", "", 300, 300, 3072, "image/png", "https://test.example.com/images/c.png")

	req := httptest.NewRequest("GET", "/api/images", nil)
	rec := httptest.NewRecorder()

	handleListImages(db)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var images []Image
	if err := json.NewDecoder(rec.Body).Decode(&images); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(images) != 3 {
		t.Errorf("expected 3 images, got %d", len(images))
	}
}

func TestHandleListImages_Search(t *testing.T) {
	db := testDB(t)

	db.Exec(`INSERT INTO images (r2_key, filename, alt_text, width, height, size_bytes, content_type, public_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"images/a.png", "photo-beach.png", "", 100, 100, 1024, "image/png", "https://test.example.com/images/a.png")
	db.Exec(`INSERT INTO images (r2_key, filename, alt_text, width, height, size_bytes, content_type, public_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"images/b.png", "photo-mountain.png", "", 200, 200, 2048, "image/png", "https://test.example.com/images/b.png")
	db.Exec(`INSERT INTO images (r2_key, filename, alt_text, width, height, size_bytes, content_type, public_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"images/c.png", "logo.png", "", 50, 50, 512, "image/png", "https://test.example.com/images/c.png")

	req := httptest.NewRequest("GET", "/api/images?search=photo", nil)
	rec := httptest.NewRecorder()

	handleListImages(db)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var images []Image
	json.NewDecoder(rec.Body).Decode(&images)

	if len(images) != 2 {
		t.Errorf("expected 2 images matching 'photo', got %d", len(images))
	}
}

func TestHandleListImages_Pagination(t *testing.T) {
	db := testDB(t)

	for i := 0; i < 5; i++ {
		db.Exec(`INSERT INTO images (r2_key, filename, alt_text, width, height, size_bytes, content_type, public_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			fmt.Sprintf("images/%d.png", i), fmt.Sprintf("img%d.png", i), "", 10, 10, 100, "image/png", fmt.Sprintf("https://test.example.com/images/%d.png", i))
	}

	// Page 1, per_page 2
	req := httptest.NewRequest("GET", "/api/images?page=1&per_page=2", nil)
	rec := httptest.NewRecorder()
	handleListImages(db)(rec, req)

	var page1 []Image
	json.NewDecoder(rec.Body).Decode(&page1)
	if len(page1) != 2 {
		t.Errorf("page 1: expected 2 images, got %d", len(page1))
	}

	// Page 3, per_page 2 (should get 1)
	req2 := httptest.NewRequest("GET", "/api/images?page=3&per_page=2", nil)
	rec2 := httptest.NewRecorder()
	handleListImages(db)(rec2, req2)

	var page3 []Image
	json.NewDecoder(rec2.Body).Decode(&page3)
	if len(page3) != 1 {
		t.Errorf("page 3: expected 1 image, got %d", len(page3))
	}
}

func TestHandleDeleteImage(t *testing.T) {
	db := testDB(t)
	var deletedKey string
	store := &mockObjectStore{
		deleteObjectFn: func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
			deletedKey = *params.Key
			return &s3.DeleteObjectOutput{}, nil
		},
	}
	cfg := testImageConfig()

	// Insert an image
	result, _ := db.Exec(`INSERT INTO images (r2_key, filename, alt_text, width, height, size_bytes, content_type, public_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"images/to-delete.png", "to-delete.png", "", 10, 10, 100, "image/png", "https://test.example.com/images/to-delete.png")
	id, _ := result.LastInsertId()

	req := httptest.NewRequest("DELETE", "/api/images/"+strconv.FormatInt(id, 10), nil)
	req.SetPathValue("id", strconv.FormatInt(id, 10))
	rec := httptest.NewRecorder()

	handleDeleteImage(db, store, cfg)(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	if deletedKey != "images/to-delete.png" {
		t.Errorf("expected deleted key 'images/to-delete.png', got %q", deletedKey)
	}

	// Verify deleted from DB
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM images WHERE id = ?`, id).Scan(&count)
	if count != 0 {
		t.Error("expected image to be deleted from DB")
	}
}

func TestHandleDeleteImage_NotFound(t *testing.T) {
	db := testDB(t)
	store := &mockObjectStore{}
	cfg := testImageConfig()

	req := httptest.NewRequest("DELETE", "/api/images/999", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()

	handleDeleteImage(db, store, cfg)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandleDeleteImage_CleansUpContentImages(t *testing.T) {
	db := testDB(t)
	store := &mockObjectStore{}
	cfg := testImageConfig()

	// Create a content entry
	contentResult, _ := db.Exec(`INSERT INTO content (type, slug, title, body) VALUES (?, ?, ?, ?)`,
		"article", "test-article", "Test Article", "body")
	contentID, _ := contentResult.LastInsertId()

	// Create an image
	imgResult, _ := db.Exec(`INSERT INTO images (r2_key, filename, alt_text, width, height, size_bytes, content_type, public_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"images/linked.png", "linked.png", "", 10, 10, 100, "image/png", "https://test.example.com/images/linked.png")
	imgID, _ := imgResult.LastInsertId()

	// Link them
	db.Exec(`INSERT INTO content_images (content_id, image_id) VALUES (?, ?)`, contentID, imgID)

	req := httptest.NewRequest("DELETE", "/api/images/"+strconv.FormatInt(imgID, 10), nil)
	req.SetPathValue("id", strconv.FormatInt(imgID, 10))
	rec := httptest.NewRecorder()

	handleDeleteImage(db, store, cfg)(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	// Verify content_images cleaned up
	var junctionCount int
	db.QueryRow(`SELECT COUNT(*) FROM content_images WHERE image_id = ?`, imgID).Scan(&junctionCount)
	if junctionCount != 0 {
		t.Error("expected content_images junction to be cleaned up")
	}
}

func TestHandleUpdateImageAlt(t *testing.T) {
	db := testDB(t)

	result, _ := db.Exec(`INSERT INTO images (r2_key, filename, alt_text, width, height, size_bytes, content_type, public_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"images/alt-test.png", "alt-test.png", "old alt", 10, 10, 100, "image/png", "https://test.example.com/images/alt-test.png")
	id, _ := result.LastInsertId()

	body, _ := json.Marshal(map[string]string{"alt_text": "new alt text"})
	req := httptest.NewRequest("PUT", "/api/images/"+strconv.FormatInt(id, 10), bytes.NewReader(body))
	req.SetPathValue("id", strconv.FormatInt(id, 10))
	rec := httptest.NewRecorder()

	handleUpdateImageAlt(db)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var img Image
	json.NewDecoder(rec.Body).Decode(&img)

	if img.AltText != "new alt text" {
		t.Errorf("expected alt_text 'new alt text', got %q", img.AltText)
	}
}

func TestHandleUpdateImageAlt_NotFound(t *testing.T) {
	db := testDB(t)

	body, _ := json.Marshal(map[string]string{"alt_text": "test"})
	req := httptest.NewRequest("PUT", "/api/images/999", bytes.NewReader(body))
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()

	handleUpdateImageAlt(db)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal.png", "normal.png"},
		{"my photo (1).png", "my_photo__1_.png"},
		{"../../../etc/passwd", "passwd"},
		{"", "upload"},
		{"héllo wörld.jpg", "h_llo_w_rld.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
