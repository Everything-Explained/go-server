package main

import (
	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/routes"
)

func main() {
	cfg := configs.GetConfig()
	r := router.NewRouter()
	routes.HandleSetup(r)
	routes.HandleRed33m(r)
	routes.HandleData(r)
	routes.HandleAssets(r)
	routes.HandleIndex(r, cfg.ClientPath+"/index.html")

	r.Listen("127.0.0.1", cfg.Port)
}
