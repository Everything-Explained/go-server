package main

import (
	"fmt"
	"net/http"

	"github.com/Everything-Explained/go-server/internal/utils"
)

func main() {
	fmt.Printf("InDev: %v\nConfigFilePath: %s\n", utils.Env.InDev, utils.Env.ConfigFilePath)
	http.ListenAndServe("127.0.0.1:8080", nil)
}
