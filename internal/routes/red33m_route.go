package routes

import (
	"fmt"
	"net/http"

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
func HandleRed33m(
	rt *router.Router,
	u *db.Users,
	password string,
	mw ...router.Middleware,
) {
	rt.Post("/red33m", func(w http.ResponseWriter, r *http.Request) {
		agData := middleware.GetAuthGuardData(r)

		if agData.IsRed33med {
			http.Error(w, "Already Logged In", http.StatusBadRequest)
			return
		}

		body := router.ReadBody(r)
		if body == "" {
			http.Error(w, "Missing Body", http.StatusBadRequest)
			return
		}

		err := bcrypt.CompareHashAndPassword(
			[]byte(password),
			[]byte(body),
		)
		if err != nil {
			http.Error(w, "Invalid Password", http.StatusUnauthorized)
			return
		}
		fmt.Fprint(w, "Login Successful\n")
		u.Update(agData.UserID, true)
	}, mw...)
}
