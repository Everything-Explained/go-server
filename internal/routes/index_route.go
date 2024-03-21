package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal/router"
)

func HandleIndex(r *router.Router, filePath string) {
	fmt.Printf("index path: %s\n", filePath)
	r.Get("/*", func(rw router.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.Path, ".") {
			rw.WriteHeader(404)
			return
		}
		err := router.FileServer.ServeNoCache(filePath, rw, req)
		if err != nil {
			panic(err)
		}
	})
}
