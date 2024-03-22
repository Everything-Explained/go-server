package guards

import (
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/writers"
)

type AuthGuard struct {
	Handler         func(rw router.ResponseWriter, req *http.Request) (string, int)
	GetContextValue func(rw router.ResponseWriter) (authGuardData, error)
}

/*
GetAuthGuard returns an authorization-guard middleware, that limits
resource access to authorized users only.

📝 Users are authorized through the '/setup' route, which is
white-listed when containing the 'Bearer setup' Authorization
header.
*/
func GetAuthGuard() AuthGuard {
	authGuardContextKey := &router.ContextKey{Name: "auth"}

	return AuthGuard{
		Handler: setupAuthGuard(authGuardContextKey),
		GetContextValue: func(rw router.ResponseWriter) (authGuardData, error) {
			return router.GetContextValue[authGuardData](authGuardContextKey, rw)
		},
	}
}

type authGuardData struct {
	IsRed33med bool
	HasAuth    bool
	Token      string
}

func setupAuthGuard(ctxKey *router.ContextKey) router.GuardFunc {
	return func(rw router.ResponseWriter, req *http.Request) (string, int) {
		isSettingUp := req.URL.Path == "/setup"
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
			return "suspicious activity detected", 403
		}

		rw.WithValue(ctxKey, authGuardData{
			IsRed33med: userState == 1,
			HasAuth:    true,
			Token:      token,
		})

		return "", 0
	}
}
