package routes

import (
	"fmt"
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
	)
}

func getSetupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeadArray := r.Header["Authorization"]
		isValidAuth := strings.Contains(strings.TrimSpace(authHeadArray[0]), " ")
		if len(authHeadArray) == 0 || !isValidAuth {
			w.WriteHeader(403)
			fmt.Fprint(w, "suspicious activity detected")
			return
		}

		authStr := strings.TrimSpace(authHeadArray[0])
		if authStr == "Bearer setup" {
			id := writers.UserWriter.AddUser(false)
			w.Header().Add("X-Evex-Id", id)
			w.Header().Add("X-Evex-Red33m", "no")
			sendVersionFile(w, r)
			return
		}

		id := strings.Split(authStr, " ")[1]
		state, err := writers.UserWriter.GetUserState(id)
		if err != nil {
			// Client should try to get a new ID
			w.WriteHeader(205)
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
