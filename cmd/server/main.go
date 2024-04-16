package main

import (
	"net/http"

	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/middleware"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/routes"
)

func main() {
	cfg := configs.GetConfig()

	rootRouter := router.NewRouter()

	assetDir := configs.GetConfig().ClientPath + "/assets"
	routes.HandleAssets(rootRouter, assetDir, middleware.LogRequests(http.StatusBadRequest))
	routes.HandleSetup(
		rootRouter,
		cfg.DataPath+"/versions.json",
		middleware.LogRequests(http.StatusBadRequest),
	)
	routes.HandleIndex(
		rootRouter,
		cfg.ClientPath+"/index.html",
		middleware.LogRequests(0),
	)

	authRouter := router.NewRouter()
	routes.HandleRed33m(authRouter)
	dataPath := configs.GetConfig().DataPath
	routes.HandleData(authRouter, dataPath)

	router.AddSubRoute(
		"/authed",
		rootRouter,
		authRouter,
		middleware.LogRequests(http.StatusBadRequest),
		middleware.AuthGuard,
	)

	err := rootRouter.ListenAndServe("127.0.0.1", cfg.Port)
	if err != nil {
		panic(err)
	}
}
