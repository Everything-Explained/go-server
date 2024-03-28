package routes

import (
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/writers"
)

/*
HandleSetup responds with the version file, route authorization
ID header, & red33m status header. The client should use the
ID, when requesting from routes protected by the auth guard
middleware.
*/
func HandleSetup(r *router.Router, mw ...router.Middleware) {
	r.Get(
		"/setup",
		getSetupHandler(),
		mw...,
	)
}

func getSetupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))

		if authHeader == "" || !strings.Contains(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if authHeader == "Bearer setup" {
			id := writers.UserWriter.AddUser(false)
			w.Header().Add("X-Evex-Id", id)
			w.Header().Add("X-Evex-Red33m", "no")
			sendVersionFile(w, r)
			return
		}

		id := strings.TrimPrefix(authHeader, "Bearer ")
		state, err := writers.UserWriter.GetUserState(id)
		if err != nil {
			// Client should try to get a new ID
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		red33mState := "no"
		if state == 1 {
			red33mState = "yes"
		}

		w.Header().Add("X-Evex-Red33m", red33mState)
		sendVersionFile(w, r)
	}
}

func sendVersionFile(w http.ResponseWriter, r *http.Request) {
	versionFile := configs.GetConfig().DataPath + "/versions.json"
	err := router.FileServer.ServeNoCache(versionFile, w, r)
	if err != nil {
		panic(err)
	}
}
