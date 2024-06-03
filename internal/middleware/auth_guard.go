package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal"
	"github.com/Everything-Explained/go-server/internal/db"
	"github.com/Everything-Explained/go-server/internal/router"
)

var AuthGuardContextKey = &internal.ContextKey{Name: "auth"}

type AuthGuardData struct {
	IsRed33med bool
	UserID     string
}

func GetAuthGuardData(r *http.Request) AuthGuardData {
	data, err := router.GetContextValue[AuthGuardData](AuthGuardContextKey, r)
	if err != nil {
		panic("missing auth guard data; did you forget to add the auth guard middleware?")
	}
	return data
}

func AuthGuard(u *db.Users) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Malformed Authorization", http.StatusUnauthorized)
				return
			}

			userID := strings.TrimPrefix(authHeader, "Bearer ")
			isRed33med, err := u.GetState(userID)
			if err != nil {
				http.Error(w, "Bad User", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), AuthGuardContextKey, AuthGuardData{
				IsRed33med: isRed33med,
				UserID:     userID,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
