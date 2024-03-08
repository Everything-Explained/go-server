package main

import (
	"fmt"

	"github.com/Everything-Explained/go-server/internal"
	"github.com/Everything-Explained/go-server/internal/lib"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/router/routes"
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
	routes.AddAPISetupRoute(r)
	r.Listen("127.0.0.1", config.Port)
}
