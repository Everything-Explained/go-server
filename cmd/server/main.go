package main

import (
	"fmt"
	"net/http"

	"github.com/Everything-Explained/go-server/internal/lib"
	"github.com/Everything-Explained/go-server/internal/utils"
)

func main() {
	config := lib.GetConfig()
	fmt.Printf(
		"InDev: %v\nConfig: %s\nPort: %d\nMail: %s\n",
		utils.Env.InDev,
		utils.Env.ConfigFilePath,
		config.Port,
		config.Mail.Host,
	)
	http.ListenAndServe("127.0.0.1:8080", nil)
}
