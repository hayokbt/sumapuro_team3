package login

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gomoku-backend/internal/domain"
	"gomoku-backend/internal/interface/httpapi/auth"
)

type authenticationStub struct {
	err error
}

func (a authenticationStub) Authenticate(context.Context, string, string) (domain.AuthenticatedUser, error) {
	return domain.AuthenticatedUser{ID: 1, Version: 1}, a.err
}

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		authentication authenticationStub
		wantStatus     int
		wantMessage    string
		wantError      string
	}{
		{
			name:           "successful login",
			method:         http.MethodPost,
			body:           `{"accountName":"player","password":"correct-password"}`,
			wantStatus:     http.StatusOK,
			wantMessage:    "ログインしました。",
			authentication: authenticationStub{},
		},
		{
			name:           "invalid credentials",
			method:         http.MethodPost,
			body:           `{"accountName":"player","password":"wrong-password"}`,
			authentication: authenticationStub{err: domain.ErrInvalidCredentials},
			wantStatus:     http.StatusUnauthorized,
			wantError:      "アカウント名またはパスワードが正しくありません。",
		},
		{
			name:           "authentication failure",
			method:         http.MethodPost,
			body:           `{"accountName":"player","password":"correct-password"}`,
			authentication: authenticationStub{err: errors.New("database unavailable")},
			wantStatus:     http.StatusInternalServerError,
			wantError:      "ログインを処理できませんでした。時間をおいて再度お試しください。",
		},
		{
			name:       "unsupported method",
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
			wantError:  "method not allowed",
		},
		{
			name:       "invalid JSON",
			method:     http.MethodPost,
			body:       `{"accountName":`,
			wantStatus: http.StatusBadRequest,
			wantError:  "リクエストの形式が正しくありません。",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/login", strings.NewReader(test.body))
			recorder := httptest.NewRecorder()
			NewHandler(test.authentication, &auth.Manager{Sessions: sessionStub{}}).ServeHTTP(recorder, request)

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			var response struct {
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response.Message != test.wantMessage {
				t.Errorf("message = %q, want %q", response.Message, test.wantMessage)
			}
			if response.Error != test.wantError {
				t.Errorf("error = %q, want %q", response.Error, test.wantError)
			}
		})
	}
}

type sessionStub struct{}

func (sessionStub) Start(context.Context, domain.AuthenticatedUser) (string, error) {
	return "jwt", nil
}
func (sessionStub) Verify(context.Context, string) (domain.AuthSession, error) {
	return domain.AuthSession{}, nil
}
func (sessionStub) Logout(context.Context, string) error { return nil }
