package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const testJWTSecret = "test-secret-key-for-testing"

func testConfig(t *testing.T) config {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt hash: %v", err)
	}
	return config{
		adminUser:     "admin",
		adminPassHash: string(hash),
		jwtSecret:     testJWTSecret,
	}
}

func TestGenerateToken(t *testing.T) {
	tokenStr, err := generateToken(testJWTSecret)
	if err != nil {
		t.Fatalf("generateToken: %v", err)
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(testJWTSecret), nil
	})
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if !token.Valid {
		t.Error("expected valid token")
	}

	sub, err := token.Claims.GetSubject()
	if err != nil {
		t.Fatalf("get subject: %v", err)
	}
	if sub != "admin" {
		t.Errorf("expected subject 'admin', got %q", sub)
	}
}

func TestHandleLogin_Success(t *testing.T) {
	cfg := testConfig(t)

	body, _ := json.Marshal(loginRequest{Username: "admin", Password: "testpass"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handleLogin(cfg)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Check cookie is set
	cookies := rec.Result().Cookies()
	var tokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "token" {
			tokenCookie = c
			break
		}
	}
	if tokenCookie == nil {
		t.Fatal("expected token cookie to be set")
	}
	if !tokenCookie.HttpOnly {
		t.Error("expected HttpOnly cookie")
	}
	if tokenCookie.Value == "" {
		t.Error("expected non-empty token value")
	}
}

func TestHandleLogin_WrongPassword(t *testing.T) {
	cfg := testConfig(t)

	body, _ := json.Marshal(loginRequest{Username: "admin", Password: "wrongpass"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handleLogin(cfg)(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestHandleLogin_WrongUsername(t *testing.T) {
	cfg := testConfig(t)

	body, _ := json.Marshal(loginRequest{Username: "notadmin", Password: "testpass"})
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handleLogin(cfg)(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestHandleLogout(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	rec := httptest.NewRecorder()

	handleLogout()(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	cookies := rec.Result().Cookies()
	var tokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "token" {
			tokenCookie = c
			break
		}
	}
	if tokenCookie == nil {
		t.Fatal("expected token cookie")
	}
	if tokenCookie.MaxAge != -1 {
		t.Errorf("expected MaxAge -1, got %d", tokenCookie.MaxAge)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	cfg := testConfig(t)

	tokenStr, _ := generateToken(testJWTSecret)

	called := false
	handler := authMiddleware(cfg, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: tokenStr})
	rec := httptest.NewRecorder()

	handler(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	cfg := testConfig(t)

	called := false
	handler := authMiddleware(cfg, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if called {
		t.Error("expected next handler NOT to be called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	cfg := testConfig(t)

	called := false
	handler := authMiddleware(cfg, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "invalid.jwt.token"})
	rec := httptest.NewRecorder()

	handler(rec, req)

	if called {
		t.Error("expected next handler NOT to be called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	cfg := testConfig(t)

	// Token signed with different secret
	tokenStr, _ := generateToken("different-secret")

	called := false
	handler := authMiddleware(cfg, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: tokenStr})
	rec := httptest.NewRecorder()

	handler(rec, req)

	if called {
		t.Error("expected next handler NOT to be called")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestHandleAuthCheck(t *testing.T) {
	cfg := testConfig(t)
	tokenStr, _ := generateToken(testJWTSecret)

	handler := authMiddleware(cfg, handleAuthCheck())

	req := httptest.NewRequest("GET", "/api/auth/check", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: tokenStr})
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]bool
	json.NewDecoder(rec.Body).Decode(&result)
	if !result["authenticated"] {
		t.Error("expected authenticated: true")
	}
}

func TestLoginFlow_EndToEnd(t *testing.T) {
	cfg := testConfig(t)

	// Login
	loginBody, _ := json.Marshal(loginRequest{Username: "admin", Password: "testpass"})
	loginReq := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(loginBody))
	loginRec := httptest.NewRecorder()
	handleLogin(cfg)(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d", loginRec.Code)
	}

	// Extract token cookie
	var tokenCookie *http.Cookie
	for _, c := range loginRec.Result().Cookies() {
		if c.Name == "token" {
			tokenCookie = c
			break
		}
	}
	if tokenCookie == nil {
		t.Fatal("no token cookie after login")
	}

	// Use token to check auth
	checkReq := httptest.NewRequest("GET", "/api/auth/check", nil)
	checkReq.AddCookie(tokenCookie)
	checkRec := httptest.NewRecorder()
	authMiddleware(cfg, handleAuthCheck())(checkRec, checkReq)

	if checkRec.Code != http.StatusOK {
		t.Errorf("auth check: expected 200, got %d", checkRec.Code)
	}
}
