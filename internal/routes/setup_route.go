package routes

import (
	"net/http"

	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/middleware"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/writers"
)

func HandleSetup(r *router.Router) {
	r.Get("/setup", getSetupHandler(), middleware.AuthGuard)
}

func getSetupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agData := middleware.GetAuthGuardData(r)
		id := agData.Id
		if !agData.HasAuth {
			id = writers.UserWriter.AddUser(false)
		}

		red33mStatus := "no"
		if agData.IsRed33med {
			red33mStatus = "yes"
		}

		if !agData.HasAuth {
			w.Header().Add("X-Evex-Id", id)
		}
		w.Header().Add("X-Evex-Red33m", red33mStatus)
		versionFile := configs.GetConfig().DataPath + "/versions.json"
		err := router.FileServer.ServeNoCache(versionFile, w, r)
		if err != nil {
			panic(err)
		}
	}
}
