package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// testApp spins up an App backed by the test database (DATABASE_URL env var).
// If DATABASE_URL is not set, all tests are skipped.
func testApp(t *testing.T) *App {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set — skipping integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("cannot connect to test db: %v", err)
	}
	if err := ensureSchema(context.Background(), pool); err != nil {
		t.Fatalf("ensureSchema: %v", err)
	}
	app, err := newApp(Config{
		Port:     "0",
		Database: dbURL,
		Issuer:   "https://agentunited.ai",
		JWTTTL:   24 * time.Hour,
		KID:      "test-kid",
	})
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}
	t.Cleanup(func() { app.db.Close() })
	return app
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewBuffer(b)
}

func decodeBody(t *testing.T, resp *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return out
}

// seedUser registers a unique user and returns their email + user_id.
func seedUser(t *testing.T, app *App) (email, userID string) {
	t.Helper()
	email = fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
	rr := httptest.NewRecorder()
	app.register(rr, httptest.NewRequest(http.MethodPost, "/api/v1/users/register",
		jsonBody(t, registerReq{Email: email, Password: "Pass1234!", DisplayName: "Test User"})))
	if rr.Code != http.StatusCreated {
		t.Fatalf("seed register failed: %d %s", rr.Code, rr.Body.String())
	}
	body := decodeBody(t, rr)
	userID, _ = body["user_id"].(string)
	return email, userID
}

// ---------- validatePassword ----------

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		pw   string
		want bool
	}{
		{"short1", false},              // 6 chars — too short
		{"Pass123", false},             // 7 chars — too short
		{"password", true},             // 8 chars letters-only — valid
		{"12345678", true},             // 8 digits — valid
		{"Pass1234", true},             // 8 chars mixed — valid
		{"Gm3s-9kPx-vR4t-hWj2", true}, // Apple Keychain format — valid
	}
	for _, tc := range cases {
		err := validatePassword(tc.pw)
		if tc.want && err != nil {
			t.Errorf("%q: expected valid, got %v", tc.pw, err)
		}
		if !tc.want && err == nil {
			t.Errorf("%q: expected invalid, got nil", tc.pw)
		}
	}
}

// ---------- register (password gate) ----------

func TestRegister_WeakPassword(t *testing.T) {
	app := testApp(t)

	// Only length < 8 is rejected; no complexity requirements.
	cases := []string{"short", "1234567", "abc"}
	for _, pw := range cases {
		rr := httptest.NewRecorder()
		app.register(rr, httptest.NewRequest(http.MethodPost, "/", jsonBody(t, map[string]string{
			"email": "x@example.com", "password": pw, "display_name": "X",
		})))
		if rr.Code != http.StatusBadRequest {
			t.Errorf("pw=%q: expected 400, got %d", pw, rr.Code)
		}
	}
}

// ---------- forgot-password ----------

func TestForgotPassword_UnknownEmail(t *testing.T) {
	app := testApp(t)
	rr := httptest.NewRecorder()
	app.forgotPassword(rr, httptest.NewRequest(http.MethodPost, "/",
		jsonBody(t, map[string]string{"email": "nobody@nowhere.com"})))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := decodeBody(t, rr)
	msg, _ := body["message"].(string)
	if msg == "" {
		t.Fatal("expected message in response")
	}
}

func TestForgotPassword_KnownEmail_TokenCreated(t *testing.T) {
	app := testApp(t)
	email, _ := seedUser(t, app)

	rr := httptest.NewRecorder()
	app.forgotPassword(rr, httptest.NewRequest(http.MethodPost, "/",
		jsonBody(t, map[string]string{"email": email})))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	body := decodeBody(t, rr)
	if _, ok := body["message"]; !ok {
		t.Fatal("expected message in response")
	}

	// Token row should exist in DB.
	var count int
	err := app.db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM password_reset_tokens WHERE user_id=(SELECT id::text FROM central_users WHERE email=$1)`,
		email).Scan(&count)
	if err != nil || count == 0 {
		t.Fatalf("expected token row in DB, count=%d err=%v", count, err)
	}
}

// ---------- reset-password ----------

func TestResetPassword_HappyPath(t *testing.T) {
	app := testApp(t)
	email, userID := seedUser(t, app)

	// Insert a valid reset token manually.
	token := "test-reset-token-" + fmt.Sprintf("%d", time.Now().UnixNano())
	_, err := app.db.Exec(context.Background(),
		`INSERT INTO password_reset_tokens (token, user_id, expires_at) VALUES ($1, $2, NOW()+INTERVAL '1 hour')`,
		token, userID)
	if err != nil {
		t.Fatalf("insert token: %v", err)
	}

	rr := httptest.NewRecorder()
	app.resetPassword(rr, httptest.NewRequest(http.MethodPost, "/",
		jsonBody(t, map[string]string{"token": token, "new_password": "NewPass99!"})))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	body := decodeBody(t, rr)
	if msg, _ := body["message"].(string); msg == "" {
		t.Fatal("expected message")
	}

	// Token should be deleted.
	var count int
	app.db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM password_reset_tokens WHERE token=$1`, token).Scan(&count)
	if count != 0 {
		t.Fatal("token should have been deleted after use")
	}

	// Should be able to log in with the new password.
	rr2 := httptest.NewRecorder()
	app.login(rr2, httptest.NewRequest(http.MethodPost, "/",
		jsonBody(t, map[string]string{"email": email, "password": "NewPass99!"})))
	if rr2.Code != http.StatusOK {
		t.Fatalf("login with new password failed: %d %s", rr2.Code, rr2.Body.String())
	}
}

func TestResetPassword_ExpiredToken(t *testing.T) {
	app := testApp(t)
	_, userID := seedUser(t, app)

	token := "expired-token-" + fmt.Sprintf("%d", time.Now().UnixNano())
	_, err := app.db.Exec(context.Background(),
		`INSERT INTO password_reset_tokens (token, user_id, expires_at) VALUES ($1, $2, NOW()-INTERVAL '1 minute')`,
		token, userID)
	if err != nil {
		t.Fatalf("insert expired token: %v", err)
	}

	rr := httptest.NewRecorder()
	app.resetPassword(rr, httptest.NewRequest(http.MethodPost, "/",
		jsonBody(t, map[string]string{"token": token, "new_password": "NewPass99!"})))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	body := decodeBody(t, rr)
	if body["error"] != "invalid_or_expired_token" {
		t.Fatalf("unexpected error: %v", body["error"])
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	app := testApp(t)
	rr := httptest.NewRecorder()
	app.resetPassword(rr, httptest.NewRequest(http.MethodPost, "/",
		jsonBody(t, map[string]string{"token": "no-such-token", "new_password": "NewPass99!"})))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	body := decodeBody(t, rr)
	if body["error"] != "invalid_or_expired_token" {
		t.Fatalf("unexpected error: %v", body["error"])
	}
}

func TestResetPassword_WeakNewPassword(t *testing.T) {
	app := testApp(t)
	rr := httptest.NewRecorder()
	app.resetPassword(rr, httptest.NewRequest(http.MethodPost, "/",
		jsonBody(t, map[string]string{"token": "any", "new_password": "weak"})))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	body := decodeBody(t, rr)
	if body["error"] != "weak_password" {
		t.Fatalf("unexpected error: %v", body["error"])
	}
}

func TestResetPassword_SingleUse(t *testing.T) {
	app := testApp(t)
	_, userID := seedUser(t, app)

	token := "single-use-token-" + fmt.Sprintf("%d", time.Now().UnixNano())
	app.db.Exec(context.Background(),
		`INSERT INTO password_reset_tokens (token, user_id, expires_at) VALUES ($1, $2, NOW()+INTERVAL '1 hour')`,
		token, userID)

	// First use succeeds.
	rr1 := httptest.NewRecorder()
	app.resetPassword(rr1, httptest.NewRequest(http.MethodPost, "/",
		jsonBody(t, map[string]string{"token": token, "new_password": "NewPass99!"})))
	if rr1.Code != http.StatusOK {
		t.Fatalf("first use failed: %d", rr1.Code)
	}

	// Second use is rejected.
	rr2 := httptest.NewRecorder()
	app.resetPassword(rr2, httptest.NewRequest(http.MethodPost, "/",
		jsonBody(t, map[string]string{"token": token, "new_password": "AnotherPass1"})))
	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("second use should fail: %d", rr2.Code)
	}
}
