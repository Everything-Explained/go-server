package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal/lib"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/router/http_interface"
)

var APIGuardData router.GuardData = router.GuardData{
	CanLog:  false,
	Handler: apiHandler,
}

func APIGuardFunc(rw *http_interface.ResponseWriter, req *http.Request) (string, int) {
	// Don't punish users for setting up authorization
	if req.URL.Path == "/api/setup" {
		return "", 0
	}

func apiGuardFunc(rw *http_interface.ResponseWriter, req *http.Request) (string, int) {
	isSettingUp := req.URL.Path == "/api/setup"
	uw := lib.UserWriter
	auth := req.Header["Authorization"]

	nonAuthedSetup := func() (string, int) {
		rw.StoreBool("isRed33med", false)
		rw.StoreBool("hasAuth", false)
		return "", 0
	}

	if len(auth) == 0 {
		if isSettingUp {
			return nonAuthedSetup()
		}
		return "missing auth", 401
	}

	if !strings.HasPrefix(auth[0], "Bearer ") {
		if isSettingUp {
			return nonAuthedSetup()
		}
		return "missing bearer", 401
	}

	token := strings.Split(auth[0], " ")[1]
	userState, err := uw.GetUserState(token)
	if err != nil {
		if isSettingUp {
			return nonAuthedSetup()
		}
		return "invalid auth", 403
	}

	rw.StoreBool("isRed33med", userState == 1)
	rw.StoreBool("hasAuth", true)
	return "", 0
}

func apiHandler(rw *http_interface.ResponseWriter, req *http.Request) {
	isRed33med := rw.GetBool("isRed33med")
	fmt.Println(isRed33med)
	rw.WriteHeader(200)
}
