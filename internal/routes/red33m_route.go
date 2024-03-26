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

func HandleRed33m(r *router.Router) {
	r.Post("/red33m", func(w http.ResponseWriter, r *http.Request) {
		agData := middleware.GetAuthGuardData(r)

		if agData.IsRed33med {
			w.WriteHeader(400)
			fmt.Fprintf(w, "already logged in")
			return
		}

		body := router.GetBody(r)
		if body == "" {
			w.WriteHeader(400)
			fmt.Fprintf(w, "missing body")
			return
		}

		err := bcrypt.CompareHashAndPassword(
			[]byte(configs.GetConfig().Red33mPassword),
			[]byte(body),
		)
		if err != nil {
			w.WriteHeader(401)
			fmt.Fprintf(w, "invalid password")
			return
		}

		writers.UserWriter.UpdateUser(agData.Id, true)
	}, middleware.AuthGuard)
}
