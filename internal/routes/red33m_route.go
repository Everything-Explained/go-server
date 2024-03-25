package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/routes/guards"
	"github.com/Everything-Explained/go-server/internal/writers"
	"golang.org/x/crypto/bcrypt"
)

func HandleRed33m(r *router.Router) {
	ag := guards.GetAuthGuard()

	r.PostWithGuard("/red33m", ag.HandlerFunc, router.RouteData{
		// PreMiddleware: []router.HandlerFunc{
		// 	middleware.LogHandler.IncomingReq,
		// },
		// PostMiddleware: []router.HandlerFunc{
		// 	middleware.LogHandler.OutgoingResp,
		// },
		HandlerFunc: getRed33mRoute(ag),
	})
}

func getRed33mRoute(ag guards.AuthGuard) router.HandlerFunc {
	return func(rw router.ResponseWriter, req *http.Request) {
		ctxVal, err := ag.GetContextValue(rw)
		if err != nil {
			panic(err)
		}

		if ctxVal.IsRed33med {
			rw.WriteHeader(400)
			fmt.Fprintf(rw, "already logged in")
			return
		}

		body := strings.TrimSpace(rw.GetBody())
		if body == "" {
			rw.WriteHeader(400)
			fmt.Fprintf(rw, "missing body")
			return
		}

		err = bcrypt.CompareHashAndPassword(
			[]byte(configs.GetConfig().Red33mPassword),
			[]byte(body),
		)
		if err != nil {
			rw.WriteHeader(401)
			fmt.Fprintf(rw, "invalid password")
			return
		}

		writers.UserWriter.UpdateUser(ctxVal.Id, true)
	}
}
