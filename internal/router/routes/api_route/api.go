package api_route

import (
	"context"
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal/lib"
	"github.com/Everything-Explained/go-server/internal/router/http_interface"
)

type apiGuardData struct {
	isRed33med bool
	hasAuth    bool
	token      string
}

type apiGuardKey string

var guardKey apiGuardKey = "api_guard"

func GetAPIGuardData(rw *http_interface.ResponseWriter) apiGuardData {
	if v := rw.GetContext().Value(guardKey); v == nil {
		panic("missing 'api_guard' data")
	} else {
		return v.(apiGuardData)
	}
}

func apiGuardFunc(rw *http_interface.ResponseWriter, req *http.Request) (string, int) {
	isSettingUp := req.URL.Path == "/api/setup"
	uw := lib.UserWriter
	auth := req.Header["Authorization"]

	if len(auth) == 0 {
		return "missing auth", 401
	}

	if !strings.HasPrefix(auth[0], "Bearer ") {
		return "missing bearer", 401
	}

	if isSettingUp && auth[0] == "Bearer setup" {
		rw.SetContext(context.WithValue(rw.GetContext(), guardKey, apiGuardData{
			isRed33med: false,
			hasAuth:    false,
		}))
		return "", 0
	}

	token := strings.Split(auth[0], " ")[1]
	userState, err := uw.GetUserState(token)
	if err != nil {
		return "invalid auth", 403
	}

	rw.SetContext(context.WithValue(rw.GetContext(), guardKey, apiGuardData{
		isRed33med: userState == 1,
		hasAuth:    true,
		token:      token,
	}))
	return "", 0
}
