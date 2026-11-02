package accountrecovery

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gomoku-backend/internal/domain/account"
	recoveryusecase "gomoku-backend/internal/usecase/account/accountrecovery"
)

type recoveryServiceStub struct {
	requestErr error
	verifyErr  error
	resetErr   error
	token      string
}

func (s recoveryServiceStub) RequestCode(context.Context, string) error {
	return s.requestErr
}

func (s recoveryServiceStub) VerifyCode(context.Context, string, string) (string, error) {
	return s.token, s.verifyErr
}

func (s recoveryServiceStub) ResetAccount(context.Context, string, string, string, string) error {
	return s.resetErr
}

func TestRecoveryHandlers(t *testing.T) {
	tests := []struct {
		name       string
		handler    func(recoveryusecase.Service) http.HandlerFunc
		body       string
		service    recoveryServiceStub
		wantStatus int
		wantText   string
	}{
		{
			name:       "request recovery code",
			handler:    NewRequestHandler,
			body:       `{"email":"player@example.com"}`,
			wantStatus: http.StatusAccepted,
			wantText:   "アカウントが存在する場合、確認コードをメールで送信しました。",
		},
		{
			name:       "verify recovery code",
			handler:    NewVerifyHandler,
			body:       `{"email":"player@example.com","code":"012345"}`,
			service:    recoveryServiceStub{token: "reset-token"},
			wantStatus: http.StatusOK,
			wantText:   "reset-token",
		},
		{
			name:       "invalid recovery code",
			handler:    NewVerifyHandler,
			body:       `{"email":"player@example.com","code":"000000"}`,
			service:    recoveryServiceStub{verifyErr: account.ErrInvalidCode},
			wantStatus: http.StatusBadRequest,
			wantText:   "確認コードが正しくありません。",
		},
		{
			name:       "reset account",
			handler:    NewResetHandler,
			body:       `{"email":"player@example.com","resetToken":"token","accountName":"new-player","password":"strong-password"}`,
			wantStatus: http.StatusOK,
			wantText:   "アカウント情報を更新しました。",
		},
		{
			name:       "duplicate account name",
			handler:    NewResetHandler,
			body:       `{"email":"player@example.com","resetToken":"token","accountName":"existing-player","password":"strong-password"}`,
			service:    recoveryServiceStub{resetErr: account.ErrAccountExists},
			wantStatus: http.StatusConflict,
			wantText:   "そのアカウント名は既に使用されています。",
		},
		{
			name:       "internal recovery error",
			handler:    NewRequestHandler,
			body:       `{"email":"player@example.com"}`,
			service:    recoveryServiceStub{requestErr: errors.New("database unavailable")},
			wantStatus: http.StatusInternalServerError,
			wantText:   "アカウント復旧を処理できませんでした。時間をおいて再度お試しください。",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			recorder := httptest.NewRecorder()
			test.handler(test.service).ServeHTTP(recorder, request)

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			var response map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response["message"] != test.wantText && response["error"] != test.wantText &&
				response["resetToken"] != test.wantText {
				t.Errorf("response = %v, want it to contain %q", response, test.wantText)
			}
		})
	}
}
