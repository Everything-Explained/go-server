package main

import (
	"fmt"
	"net/http"

	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/middleware"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/routes"
)

func main() {
	cfg := configs.GetConfig()

	r := router.NewRouter()
	routes.HandleAssets(r, middleware.LogRequests(http.StatusBadRequest))
	routes.HandleSetup(r, middleware.LogRequests(http.StatusBadRequest))
	routes.HandleRed33m(r, middleware.LogRequests(0), middleware.AuthGuard)
	routes.HandleData(
		r,
		middleware.LogRequests(http.StatusBadRequest),
		middleware.AuthGuard,
	)
	routes.HandleIndex(
		r,
		cfg.ClientPath+"/index.html",
		middleware.LogRequests(http.StatusBadRequest),
	)

	s := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r.Handler,
	}

	s.ListenAndServe()
}
