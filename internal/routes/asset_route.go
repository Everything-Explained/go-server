package routes

import (
	"github.com/Everything-Explained/go-server/internal/router"
)

func HandleAssets(r *router.Router, dir string, mw ...router.Middleware) {
	r.SetStaticRoute("/assets", dir, mw...)
}
