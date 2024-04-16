package routes

import (
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal/db"
	"github.com/Everything-Explained/go-server/internal/router"
)

/*
HandleSetup responds with the version file (vFilePath), route authorization
ID header, & red33m status header. The client should use the ID, when
requesting from routes protected by the auth guard middleware.
*/
func HandleSetup(r *router.Router, vFilePath string, mw ...router.Middleware) {
	r.Get(
		"/setup",
		getSetupHandler(vFilePath),
		mw...,
	)
}

func getSetupHandler(vFilePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))

		if authHeader == "" || !strings.Contains(authHeader, "Bearer ") {
			http.Error(w, "Malformed Authorization", http.StatusForbidden)
			return
		}

		users := db.GetUsers()

		if authHeader == "Bearer setup" {
			id := users.Add(false)
			w.Header().Add("X-Evex-Id", id)
			w.Header().Add("X-Evex-Red33m", "no")
			sendVersionFile(w, r, vFilePath)
			return
		}

		id := strings.TrimPrefix(authHeader, "Bearer ")
		state, err := users.GetState(id)
		if err != nil {
			// Client should try to get a new ID
			http.Error(w, "Authorization Expired or Missing", http.StatusUnauthorized)
			return
		}

		red33mState := "no"
		if state {
			red33mState = "yes"
		}

		w.Header().Add("X-Evex-Red33m", red33mState)
		sendVersionFile(w, r, vFilePath)
	}
}

// TODO  Return error and log it as server error
func sendVersionFile(w http.ResponseWriter, r *http.Request, vFilePath string) {
	err := router.FileServer.ServeFile(vFilePath, w, r, false)
	if err != nil {
		panic(err)
	}
}
