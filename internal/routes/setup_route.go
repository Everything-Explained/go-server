package routes

import (
	"net/http"

	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/routes/guards"
	"github.com/Everything-Explained/go-server/internal/writers"
)

func HandleSetup(r *router.Router) {
	ag := guards.GetAuthGuard(guards.AuthWithSetupRoute("/setup"))
	r.AddGetGuard("/setup", ag.Handler, router.RouteData{
		Handler: getSetupRoute(ag),
	})
}

func getSetupRoute(ag guards.AuthGuard) router.HandlerFunc {
	return func(rw router.ResponseWriter, req *http.Request) {
		ctx, err := ag.GetContextValue(rw)
		if err != nil {
			panic(err)
		}
		token := ctx.Token
		if !ctx.HasAuth {
			token = writers.UserWriter.AddUser(false)
		}

		red33mStatus := "no"
		if ctx.IsRed33med {
			red33mStatus = "yes"
		}

		if !ctx.HasAuth {
			rw.Header().Add("X-Evex-Token", token)
		}
		rw.Header().Add("X-Evex-Red33m", red33mStatus)
		versionFile := configs.GetConfig().DataPath + "/versions.json"
		err = router.FileServer.ServeNoCache(versionFile, rw, req)
		if err != nil {
			panic(err)
		}
	}
}
