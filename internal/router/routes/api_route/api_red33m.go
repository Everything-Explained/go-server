package api_route

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Everything-Explained/go-server/internal/lib"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/router/http_interface"
	"github.com/alexedwards/argon2id"
)

func AddRed33mRoute(r *router.Router) {
	r.AddPostGuard("/api/red33m", apiGuardFunc, router.GuardData{
		CanLog:  true,
		Handler: red33mRoute,
	})
}

func red33mRoute(rw *http_interface.ResponseWriter, req *http.Request) {
	guardData := GetAPIGuardData(rw)
	if guardData.isRed33med {
		rw.WriteHeader(400)
		fmt.Fprint(rw, "already logged in")
		return
	}

	body := rw.GetBody()
	if body == "" {
		rw.WriteHeader(400)
		fmt.Fprint(rw, "missing body")
		return
	}

	isValid, err := argon2id.ComparePasswordAndHash(body, lib.GetConfig().Red33mPassword)
	if err != nil {
		// todo - log to server error log
		log.Fatal(err)
		rw.WriteHeader(500)
		return
	}

	if !isValid {
		rw.WriteHeader(401)
		fmt.Fprint(rw, "invalid passcode")
		return
	}

	rw.WriteHeader(200)
}
