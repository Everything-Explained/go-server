package routes

import (
	"net/http"

	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/routes/guards"
	"github.com/Everything-Explained/go-server/internal/writers"
)

func HandleSetup(r *router.Router) {
	ag := guards.GetAuthGuard()
	r.AddGetGuard("/setup", ag.Handler, router.RouteData{
		Handler: getSetupRoute(ag),
	})
}

func getSetupRoute(ag guards.AuthGuard) router.HandlerFunc {
	return func(rw router.ResponseWriter, req *http.Request) {
		ctxVal, err := ag.GetContextValue(rw)
		if err != nil {
			panic(err)
		}
		id := ctxVal.Id
		if !ctxVal.HasAuth {
			id = writers.UserWriter.AddUser(false)
		}

		red33mStatus := "no"
		if ctxVal.IsRed33med {
			red33mStatus = "yes"
		}

		if !ctxVal.HasAuth {
			rw.Header().Add("X-Evex-Id", id)
		}
		rw.Header().Add("X-Evex-Red33m", red33mStatus)
		versionFile := configs.GetConfig().DataPath + "/versions.json"
		err = router.FileServer.ServeNoCache(versionFile, rw, req)
		if err != nil {
			panic(err)
		}
	}
}
