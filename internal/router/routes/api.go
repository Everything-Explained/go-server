package routes

import (
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal/lib"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/router/http_interface"
)

func AddAPISetupRoute(r *router.Router) {
	r.AddGetGuard("/api/setup", apiGuardFunc, router.GuardData{
		CanLog:  true,
		Handler: setupRoute,
	})
}

func setupRoute(rw *http_interface.ResponseWriter, req *http.Request) {
	token := rw.GetStr("token")
	if !rw.GetBool("hasAuth") {
		token = lib.UserWriter.AddUser(false)
	}

	red33mStatus := "no"
	if rw.GetBool("isRed33med") {
		red33mStatus = "yes"
	}

	rw.Header().Add("X-Evex-Token", token)
	rw.Header().Add("X-Evex-Red33m", red33mStatus)
	// todo - send versions.json file
	rw.WriteHeader(200)
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
