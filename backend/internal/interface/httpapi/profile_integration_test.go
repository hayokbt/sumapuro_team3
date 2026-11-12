package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gomoku-backend/internal/infrastructure/database"
	"gomoku-backend/internal/infrastructure/password"
	"gomoku-backend/internal/infrastructure/token"
	"gomoku-backend/internal/interface/account/auth"
	"gomoku-backend/internal/interface/httpapi"
	"gomoku-backend/internal/usecase/account/authentication"
	"gomoku-backend/internal/usecase/account/authsession"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestProfileAPIIntegration(t *testing.T) {
	if os.Getenv("DB_INTEGRATION") != "1" {
		t.Skip("set DB_INTEGRATION=1 with isolated MySQL")
	}
	db, err := database.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	hasher := password.NewPBKDF2()
	hash, err := hasher.Hash("password123")
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("profile_%d", time.Now().UnixNano())
	res, err := db.Exec(`INSERT INTO users(username,email,password_hash)VALUES(?,?,?)`, name, name+"@example.com", hash)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := res.LastInsertId()
	defer db.Exec(`DELETE FROM users WHERE id=?`, id)
	codec, err := token.NewJWT("integration-profile-secret-32-bytes!")
	if err != nil {
		t.Fatal(err)
	}
	manager := &auth.Manager{Sessions: authsession.New(database.NewAuthSessionRepository(db), codec)}
	handler := httpapi.NewHandler(nil, authentication.NewAuthentication(database.NewAuthenticationRepository(db), hasher), nil, manager, database.NewRankingRepository(db), []string{"http://localhost:5173"}, database.NewProfileRepository(db), database.NewGameResultRepository(db))
	request := func(method, path string, body any, cookie *http.Cookie, csrf bool, status int) *httptest.ResponseRecorder {
		t.Helper()
		data, _ := json.Marshal(body)
		r := httptest.NewRequest(method, path, bytes.NewReader(data))
		r.Header.Set("Origin", "http://localhost:5173")
		if csrf {
			r.Header.Set("X-Requested-With", "gomoku-web")
		}
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		return w
	}
	login := func(accountName, password string) *http.Cookie {
		t.Helper()
		w := request("POST", "/login", map[string]string{"accountName": accountName, "password": password}, nil, true, 200)
		return w.Result().Cookies()[0]
	}
	first, second := login(name, "password123"), login(name, "password123")
	request("GET", "/account", nil, nil, true, 401)
	request("POST", "/account/settings", map[string]int{"bgmVolume": 0, "seVolume": 100}, first, false, 403)
	request("POST", "/account/settings", map[string]int{"bgmVolume": 0, "seVolume": 100}, first, true, 200)
	w := request("GET", "/account/settings", nil, second, true, 200)
	var settings map[string]int
	if err = json.Unmarshal(w.Body.Bytes(), &settings); err != nil || settings["bgmVolume"] != 0 || settings["seVolume"] != 100 {
		t.Fatalf("settings %s %v", w.Body.String(), err)
	}
	request("POST", "/account", map[string]string{"accountName": name + "_new", "currentPassword": "password123", "password": "new-password123"}, first, true, 200)
	request("GET", "/account", nil, first, true, 401)
	request("GET", "/account", nil, second, true, 401)
	fresh := login(name+"_new", "new-password123")
	request("GET", "/account", nil, fresh, true, 200)
}
