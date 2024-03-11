package routes

import (
	"fmt"
	"log"
	"net/http"
	"os"
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
		ff, err := lib.FastFileServer(indexPath, "")
		if err != nil {
			if os.IsNotExist(err) {
				log.Fatal("missing index page")
				rw.WriteHeader(500)
				return
			}
			panic(err)
		}
		rw.Header().Add("Content-Type", ff.ContentType)
		rw.Header().Add("Content-Length", fmt.Sprintf("%d", ff.Length))
		rw.Write(ff.Content)
	})
}
