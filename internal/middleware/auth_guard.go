package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/writers"
)

var AuthGuardContextKey = &internal.ContextKey{Name: "auth"}

type AuthGuardData struct {
	IsRed33med bool
	HasAuth    bool
	Id         string
}

func GetAuthGuardData(r *http.Request) AuthGuardData {
	data, err := router.GetContextValue[AuthGuardData](AuthGuardContextKey, r)
	if err != nil {
		panic("missing auth guard data; did you forget to add the auth guard middleware?")
	}
	return data
}

func AuthGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "missing auth")
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "missing bearer")
			return
		}

		id := strings.TrimPrefix(authHeader, "Bearer ")
		userState, err := writers.UserWriter.GetUserState(id)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, "suspicious activity detected")
			return
		}

		ctx := context.WithValue(r.Context(), AuthGuardContextKey, AuthGuardData{
			IsRed33med: userState == 1,
			HasAuth:    true,
			Id:         id,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
