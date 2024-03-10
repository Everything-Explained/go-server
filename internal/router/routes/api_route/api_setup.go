package api_route

import (
	"net/http"

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
