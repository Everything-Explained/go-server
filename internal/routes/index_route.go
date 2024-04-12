package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal/router"
)

func HandleIndex(r *router.Router, filePath string, mw ...router.Middleware) {
	fmt.Printf("index path: %s\n", filePath)
	if !strings.Contains(filePath, ".") {
		panic("index route needs a file path, not folder path")
	}
	r.Any("/", func(rw http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.Path, ".") && req.URL.Path != "/index.html" {
			http.Error(rw, "Page Not Found", http.StatusNotFound)
			return
		}
		err := router.FileServer.ServeFile(filePath, rw, req, false)
		if err != nil {
			panic(err)
		}
	}, mw...)
}
