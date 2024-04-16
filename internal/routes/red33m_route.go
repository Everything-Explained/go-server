package routes

import (
	"net/http"

	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/db"
	"github.com/Everything-Explained/go-server/internal/middleware"
	"github.com/Everything-Explained/go-server/internal/router"
	"golang.org/x/crypto/bcrypt"
)

/*
HandleRed33m sets the state of a user ID, to be able to access
red33m content.

🟠 Requires the auth guard middleware.
*/
func HandleRed33m(rt *router.Router, mw ...router.Middleware) {
	rt.Post("/red33m", func(w http.ResponseWriter, r *http.Request) {
		agData := middleware.GetAuthGuardData(r)

		if agData.IsRed33med {
			http.Error(w, "already logged in", http.StatusBadRequest)
			return
		}

		body := router.ReadBody(r)
		if body == "" {
			http.Error(w, "missing body", http.StatusBadRequest)
			return
		}

		err := bcrypt.CompareHashAndPassword(
			[]byte(configs.GetConfig().Red33mPassword),
			[]byte(body),
		)
		if err != nil {
			http.Error(w, "invalid password", http.StatusUnauthorized)
			return
		}
		db.GetUsers().Update(agData.Id, true)
	}, mw...)
}
