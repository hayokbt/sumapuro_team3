package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"time"

	"gomoku-backend/internal/infrastructure/database"
	"gomoku-backend/internal/infrastructure/email"
	"gomoku-backend/internal/infrastructure/password"
	"gomoku-backend/internal/infrastructure/token"
	"gomoku-backend/internal/interface/httpapi"
	"gomoku-backend/internal/interface/httpapi/auth"
	"gomoku-backend/internal/usecase/accountrecovery"
	"gomoku-backend/internal/usecase/accountregistration"
	"gomoku-backend/internal/usecase/authentication"
	"gomoku-backend/internal/usecase/authsession"
)

func TestAuthenticationIntegration(t *testing.T) {
	if os.Getenv("AUTH_INTEGRATION") != "1" {
		t.Skip("set AUTH_INTEGRATION=1 with local MySQL and Mailpit")
	}
	db, err := database.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	hasher := password.NewPBKDF2()
	hash, err := hasher.Hash("test-password-123")
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("jwt_test_%d", time.Now().UnixNano())
	address := name + "@example.com"
	result, err := db.Exec("INSERT INTO users(username,email,password_hash) VALUES(?,?,?)", name, address, hash)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := result.LastInsertId()
	defer db.Exec("DELETE FROM users WHERE id=?", id)
	mailer, err := email.NewSMTP("localhost", "1025", "no-reply@example.com")
	if err != nil {
		t.Fatal(err)
	}
	codec, _ := token.NewJWT("integration-test-key-32-bytes-long!")
	manager := &auth.Manager{Sessions: authsession.New(database.NewAuthSessionRepository(db), codec)}
	server := httpapi.NewHandler(accountregistration.NewAccountRegistration(database.NewRegistrationRepository(db), mailer, hasher), authentication.NewAuthentication(database.NewAuthenticationRepository(db), hasher), accountrecovery.NewAccountRecovery(database.NewAccountRecoveryRepository(db), mailer, hasher), manager, []string{"http://localhost:5173"})
	request := func(method, path string, body any, cookie *http.Cookie, expected int) *httptest.ResponseRecorder {
		t.Helper()
		data, _ := json.Marshal(body)
		r := httptest.NewRequest(method, path, bytes.NewReader(data))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Requested-With", "gomoku-web")
		r.Header.Set("Origin", "http://localhost:5173")
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		if w.Code != expected {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		return w
	}
	login := func(account string) *http.Cookie {
		t.Helper()
		w := request("POST", "/login", map[string]string{"accountName": account, "password": "test-password-123"}, nil, 200)
		cookies := w.Result().Cookies()
		if len(cookies) != 1 || !cookies[0].HttpOnly {
			t.Fatal("missing HttpOnly auth cookie")
		}
		return cookies[0]
	}
	request("GET", "/auth/me", nil, nil, 401)
	first := login(name)
	second := login(name)
	request("GET", "/auth/me", nil, first, 200)
	request("POST", "/logout", nil, first, 200)
	request("GET", "/auth/me", nil, first, 401)
	request("GET", "/auth/me", nil, second, 200)
	request("POST", "/account/recovery/request", map[string]string{"email": address}, nil, 202)
	var messages struct {
		Messages []struct {
			ID string
			To []struct{ Address string }
		}
	}
	readJSON := func(url string, out any) {
		t.Helper()
		resp, err := http.Get(url)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatal(err)
		}
	}
	readJSON("http://localhost:8025/api/v1/messages", &messages)
	var messageID string
	for _, message := range messages.Messages {
		for _, to := range message.To {
			if to.Address == address {
				messageID = message.ID
			}
		}
	}
	if messageID == "" {
		t.Fatal("recovery email not found")
	}
	defer func() {
		r, _ := http.NewRequest("DELETE", "http://localhost:8025/api/v1/messages", bytes.NewBufferString(`{"IDs":["`+messageID+`"]}`))
		r.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(r)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()
	var message struct{ Text string }
	readJSON("http://localhost:8025/api/v1/message/"+messageID, &message)
	code := regexp.MustCompile(`\b[0-9]{6}\b`).FindString(message.Text)
	if code == "" {
		t.Fatal("verification code missing")
	}
	verified := request("POST", "/account/recovery/verify", map[string]string{"email": address, "code": code}, nil, 200)
	var payload struct{ ResetToken string }
	json.Unmarshal(verified.Body.Bytes(), &payload)
	reset := map[string]string{"email": address, "resetToken": payload.ResetToken, "accountName": name + "_new", "password": "test-password-123"}
	request("POST", "/account/recovery/reset", reset, nil, 200)
	request("GET", "/auth/me", nil, second, 401)
	request("POST", "/account/recovery/reset", reset, nil, 410) // consumed reset token returns 410
	fresh := login(name + "_new")
	request("GET", "/auth/me", nil, fresh, 200)
	r := httptest.NewRequest("POST", "/logout", nil)
	r.AddCookie(fresh)
	r.Header.Set("Origin", "https://evil.example")
	r.Header.Set("X-Requested-With", "gomoku-web")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("cross-origin logout accepted")
	}
}
