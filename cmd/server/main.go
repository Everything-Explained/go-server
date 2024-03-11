package main

import (
	"fmt"

	"github.com/Everything-Explained/go-server/internal"
	"github.com/Everything-Explained/go-server/internal/lib"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/router/routes"
	"github.com/Everything-Explained/go-server/internal/router/routes/api_route"
)

func main() {
	config := lib.GetConfig()
	fmt.Printf(
		"InDev: %v\nConfig: %s\nPort: %d\nMail: %s\n",
		internal.GetEnv().InDev,
		internal.GetEnv().ConfigFilePath,
		config.Port,
		config.Mail.Host,
	)
	r := router.NewRouter()
	r.AddStaticRoute("/assets", "assets")
	api_route.AddAPIDataRoute(r)
	api_route.AddAPISetupRoute(r)
	api_route.AddRed33mRoute(r)
	routes.AddSPARoute(r)
	r.Listen("127.0.0.1", config.Port)
}
