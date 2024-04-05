package routes

import (
	"fmt"
	"net/http"

	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/middleware"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/writers"
	"golang.org/x/crypto/bcrypt"
)

/*
HandleRed33m sets the state of a user ID, to be able to access
red33m content.

🟠 Requires the auth guard middleware.
*/
func HandleRed33m(r *router.Router, mw ...router.Middleware) {
	r.Post("/red33m", func(w http.ResponseWriter, r *http.Request) {
		agData := middleware.GetAuthGuardData(r)

		if agData.IsRed33med {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "already logged in")
			return
		}

		body := router.ReadBody(r)
		if body == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "missing body")
			return
		}

		err := bcrypt.CompareHashAndPassword(
			[]byte(configs.GetConfig().Red33mPassword),
			[]byte(body),
		)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "invalid password")
			return
		}
		writers.UserWriter.UpdateUser(agData.Id, true)
	}, mw...)
}
