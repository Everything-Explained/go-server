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
	r.Get("/", func(rw http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.Path, ".") {
			rw.WriteHeader(404)
			return
		}
		err := router.FileServer.ServeNoCache(filePath, rw, req)
		if err != nil {
			panic(err)
		}
	}, mw...)
}
