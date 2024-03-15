package api_route

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal/lib"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/router/http_interface"
)

var dataPath string = lib.GetConfig().DataPath

func AddAPIDataRoute(r *router.Router) {
	r.AddGetGuard("/api/data/{content}/{visibility}", apiGuardFunc, router.GuardData{
		CanLog:  true,
		Handler: dataPreviewHandler,
	})
	r.AddGetGuard("/api/data/{content}/{visibility}/{file}", apiGuardFunc, router.GuardData{
		Handler: mdhtmlHandler,
	})
}

/*
mdhtmlHandler handles loading our custom MDHTML files, which contain
HTML built from markdown (hence the name).
*/
func mdhtmlHandler(rw *http_interface.ResponseWriter, req *http.Request) {
	guardData := GetAPIGuardData(rw)
	content := req.PathValue("content")
	visibility := req.PathValue("visibility")
	file := req.PathValue("file")
	notRed33med := strings.Contains(visibility, "red33m") && !guardData.isRed33med

	// Only supports MDHTML files; all other requests are abnormal
	if !strings.HasSuffix(file, ".mdhtml") || notRed33med {
		rw.WriteHeader(403)
		fmt.Fprintf(rw, "suspicious activity")
		return
	}

	filePath := fmt.Sprintf("%s/%s/%s/%s", dataPath, content, visibility, file)
	err := lib.FastFileServer.ServeMaxCache(filePath, rw, req)
	if err != nil {
		panic(err)
	}
}

/*
dataPreviewHandler loads the preview data for the available articles
at the requested route.
*/
func dataPreviewHandler(rw *http_interface.ResponseWriter, req *http.Request) {
	guardData := GetAPIGuardData(rw)
	content := req.PathValue("content")
	visibility := req.PathValue("visibility")
	notRed33med := strings.Contains(visibility, "red33m") && !guardData.isRed33med

	// Only supports non-file requests; all other requests are abnormal
	if strings.Contains(visibility, ".") || notRed33med {
		rw.WriteHeader(403)
		fmt.Fprintf(rw, "suspicious activity detected")
		return
	}

	filePath := fmt.Sprintf("%s/%s/%s/%s.json", dataPath, content, visibility, visibility)
	err := lib.FastFileServer.ServeMaxCache(filePath, rw, req)
	if err != nil {
		panic(err)
	}
}
