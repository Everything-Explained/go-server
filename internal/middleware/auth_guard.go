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
		panic(err)
	}
	return data
}

func AuthGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isSettingUp := r.URL.Path == "/setup"

		authHeadArray := r.Header["Authorization"]
		if len(authHeadArray) == 0 {
			w.WriteHeader(401)
			fmt.Fprint(w, "missing auth")
			return
		}

		authHead := strings.TrimSpace(authHeadArray[0])

		if !strings.HasPrefix(authHead, "Bearer ") {
			w.WriteHeader(401)
			fmt.Fprintf(w, "missing bearer")
			return
		}

		if isSettingUp && authHead == "Bearer setup" {
			ctx := context.WithValue(r.Context(), AuthGuardContextKey, AuthGuardData{
				IsRed33med: false,
				HasAuth:    false,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		id := strings.Split(authHead, " ")[1]
		userState, err := writers.UserWriter.GetUserState(id)
		if err != nil {
			w.WriteHeader(403)
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