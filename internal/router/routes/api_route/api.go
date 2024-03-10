package api_route

import (
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal/lib"
	"github.com/Everything-Explained/go-server/internal/router/http_interface"
)

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
		rw.StoreBool("isRed33med", false)
		rw.StoreBool("hasAuth", false)
		return "", 0
	}

	token := strings.Split(auth[0], " ")[1]
	userState, err := uw.GetUserState(token)
	if err != nil {
		return "invalid auth", 403
	}

	rw.StoreBool("isRed33med", userState == 1)
	rw.StoreBool("hasAuth", true)
	rw.StoreStr("token", token)
	return "", 0
}
