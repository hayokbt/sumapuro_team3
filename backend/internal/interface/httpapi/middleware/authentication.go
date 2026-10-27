package middleware

import (
	"context"
	"errors"
	"net/http"

	"gomoku-backend/internal/domain"
	"gomoku-backend/internal/interface/httpapi/auth"
)

type userKey struct{}

func UserID(ctx context.Context) (int64, bool) { id, ok := ctx.Value(userKey{}).(int64); return id, ok }
func Authentication(manager *auth.Manager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		session, err := manager.Sessions.Verify(r.Context(), auth.Raw(r))
		if err != nil {
			status := 500
			if errors.Is(err, domain.ErrUnauthenticated) {
				status = 401
			}
			http.Error(w, "認証が必要です。", status)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, session.UserID)))
	})
}
