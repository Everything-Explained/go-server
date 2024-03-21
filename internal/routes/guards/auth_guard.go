package guards

import (
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/writers"
)

type authGuardData struct {
	IsRed33med bool
	HasAuth    bool
	Token      string
}

type AuthGuard struct {
	Handler         func(rw router.ResponseWriter, req *http.Request) (string, int)
	GetContextValue func(rw router.ResponseWriter) (authGuardData, error)
}

type AuthGuardConfig struct {
	setupRoute string
}

type AuthGuardOption func(*AuthGuardConfig)

func AuthWithSetupRoute(route string) AuthGuardOption {
	return func(cfg *AuthGuardConfig) {
		cfg.setupRoute = route
	}
}

func GetAuthGuard(options ...AuthGuardOption) AuthGuard {
	authGuardContextKey := &router.ContextKey{Name: "auth"}
	cfg := &AuthGuardConfig{}

	for _, o := range options {
		o(cfg)
	}

	return AuthGuard{
		Handler: setupAuthGuard(authGuardContextKey, cfg.setupRoute),
		GetContextValue: func(rw router.ResponseWriter) (authGuardData, error) {
			return router.GetContextValue[authGuardData](authGuardContextKey, rw)
		},
	}
}

func setupAuthGuard(ctxKey *router.ContextKey, setupRoute string) router.GuardFunc {
	return func(rw router.ResponseWriter, req *http.Request) (string, int) {
		isSettingUp := req.URL.Path == setupRoute
		uw := writers.UserWriter
		_ = uw
		authHeadArray := req.Header["Authorization"]

		if len(authHeadArray) == 0 {
			return "missing auth", 401
		}

		authHead := strings.TrimSpace(authHeadArray[0])

		if !strings.HasPrefix(authHead, "Bearer ") {
			return "missing bearer", 401
		}

		if isSettingUp && authHead == "Bearer setup" {
			rw.WithValue(ctxKey, authGuardData{
				IsRed33med: false,
				HasAuth:    false,
			})
			return "", 0
		}

		token := strings.Split(authHead, " ")[1]
		userState, err := uw.GetUserState(token)
		if err != nil {
			return "invalid auth", 403
		}

		rw.WithValue(ctxKey, authGuardData{
			IsRed33med: userState == 1,
			HasAuth:    true,
			Token:      token,
		})

		return "", 0
	}
}
