package routes

import (
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal/lib"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/router/http_interface"
)

var indexPath string = lib.GetConfig().ClientPath + "/index.html"

/*
AddSPARoute adds the main route for redirecting all SPA requests to the index
page, excluding file requests.
*/
func AddSPARoute(r *router.Router) {
	r.Get("/*", func(rw *http_interface.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.Path, ".") {
			rw.WriteHeader(404)
			return
		}
		err := lib.FastFileServer.ServeNoCache(indexPath, rw, req)
		if err != nil {
			panic(err)
		}
	})
}
