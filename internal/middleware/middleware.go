package middleware

import (
	"context"
	"net/http"

	"sewing-ecosystem/internal/auth"
)

type ctxKey string

const UserKey ctxKey = "user"

type UserSession struct {
	UserID int64
	Role   string
}

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err == nil && cookie.Value != "" {
				claims, err := auth.ParseToken(secret, cookie.Value)
				if err == nil {
					ctx := context.WithValue(r.Context(), UserKey, UserSession{UserID: claims.UserID, Role: claims.Role})
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := CurrentUser(r); !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := CurrentUser(r)
			if !ok || u.Role != role {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func CurrentUser(r *http.Request) (UserSession, bool) {
	u, ok := r.Context().Value(UserKey).(UserSession)
	return u, ok
}
