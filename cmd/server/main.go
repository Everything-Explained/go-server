package main

import (
	"fmt"
	"net/http"

	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/routes"
)

func main() {
	cfg := configs.GetConfig()
	r := router.NewRouter()
	routes.HandleAssets(r)
	routes.HandleData(r)
	routes.HandleRed33m(r)
	routes.HandleSetup(r)
	routes.HandleIndex(r, cfg.ClientPath+"/index.html")

	s := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r.Handler,
	}

	s.ListenAndServe()
}
