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

	mainRouter := router.NewRouter()
	routes.HandleAssets(mainRouter, middleware.LogRequests(http.StatusBadRequest))
	routes.HandleSetup(mainRouter, middleware.LogRequests(http.StatusBadRequest))
	routes.HandleRed33m(mainRouter, middleware.LogRequests(0), middleware.AuthGuard)
	routes.HandleData(
		mainRouter,
		middleware.LogRequests(http.StatusBadRequest),
		middleware.AuthGuard,
	)
	routes.HandleIndex(
		mainRouter,
		cfg.ClientPath+"/index.html",
		middleware.LogRequests(http.StatusBadRequest),
	)

	s := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mainRouter.Handler,
	}

	s.ListenAndServe()
}
