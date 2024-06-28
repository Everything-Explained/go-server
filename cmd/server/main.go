package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal"
	"github.com/Everything-Explained/go-server/internal/db"
	"github.com/Everything-Explained/go-server/internal/middleware"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/routes"
)

func main() {
	cfg, err := configs.GetConfig("./configs")
	if err != nil {
		panic(err)
	}

	u, err := db.NewUsers(internal.Getwd())
	if err != nil {
		panic(err)
	}

	rootRouter := router.NewRouter()

	closeLog, logByStatus := middleware.LogRequests(internal.Getwd()+"/logs", "requests")
	defer closeLog()

	routes.HandleAssets(
		rootRouter,
		cfg.ClientPath+"/assets",
		logByStatus(http.StatusBadRequest),
	)

	routes.HandleSetup(
		rootRouter,
		cfg.DataPath+"/versions.json",
		u,
		logByStatus(http.StatusBadRequest),
	)

	routes.HandleIndex(
		rootRouter,
		cfg.ClientPath+"/index.html",
		logByStatus(http.StatusBadRequest),
	)

	authRouter := router.NewRouter()
	routes.HandleRed33m(authRouter, u, cfg.Red33mPassword)
	routes.HandleData(authRouter, cfg.DataPath)

	router.AddSubRoute(
		"/authed",
		rootRouter,
		authRouter,
		logByStatus(http.StatusBadRequest),
		middleware.AuthGuard(u),
	)

	server := rootRouter.SetupServer("127.0.0.1", cfg.Port)

	go func() {
		fmt.Printf("Listening on: http://127.0.0.1:%d\n", cfg.Port)
		err = server.ListenAndServe()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("[SIGINT] Shutting Down...")

	err = server.Shutdown(context.Background())
	if err != nil {
		panic(err)
	}
}
