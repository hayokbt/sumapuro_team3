package middleware

import (
	"net/http"
	"net/url"
)

// Browser mutations must use JSON/custom headers and an explicitly trusted origin.
func CSRF(origins []string, next http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, o := range origins {
		allowed[o] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
			origin := r.Header.Get("Origin")
			if origin != "" {
				u, err := url.Parse(origin)
				if err != nil || u.Host == "" || !allowed[origin] {
					http.Error(w, "許可されていない送信元です。", 403)
					return
				}
			}
			if r.Header.Get("X-Requested-With") != "gomoku-web" {
				http.Error(w, "確認ヘッダーが必要です。", 403)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
