package api_route

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Everything-Explained/go-server/internal/lib"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/router/http_interface"
)

var (
	dataPath string = lib.GetConfig().DataPath
	maxAge   int    = 60 * 60 * 24 * 365 * 10 // 10 years
)

func AddAPIDataRoute(r *router.Router) {
	r.AddGetGuard("/api/data/{content}/{visibility}", apiGuardFunc, router.GuardData{
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
	content := req.PathValue("content")
	visibility := req.PathValue("visibility")
	file := req.PathValue("file")
	notRed33med := strings.Contains(visibility, "red33m") && !rw.GetBool("isRed33med")

	// Only supports MDHTML files; all other requests are abnormal
	if !strings.HasSuffix(file, ".mdhtml") || notRed33med {
		rw.WriteHeader(403)
		fmt.Fprintf(rw, "suspicious activity")
		return
	}

	filePath := fmt.Sprintf("%s/%s/%s/%s", dataPath, content, visibility, file)
	useFastFileServer(filePath, rw)
}

/*
dataPreviewHandler loads the preview data for the available articles
at the requested route.
*/
func dataPreviewHandler(rw *http_interface.ResponseWriter, req *http.Request) {
	content := req.PathValue("content")
	visibility := req.PathValue("visibility")
	notRed33med := strings.Contains(visibility, "red33m") && !rw.GetBool("isRed33med")

	// Only supports non-file requests; all other requests are abnormal
	if strings.Contains(visibility, ".") || notRed33med {
		rw.WriteHeader(403)
		fmt.Fprintf(rw, "suspicious activity detected")
		return
	}

	filePath := fmt.Sprintf("%s/%s/%s/%s.json", dataPath, content, visibility, visibility)
	useFastFileServer(filePath, rw)
}

func useFastFileServer(filePath string, rw *http_interface.ResponseWriter) {
	ff, err := lib.FastFileServer(filePath, "")
	if err != nil {
		if os.IsNotExist(err) {
			rw.WriteHeader(404)
			return
		}
		panic(err)
	}
	rw.Header().Add("Content-Type", ff.ContentType)
	rw.Header().Add("Content-Length", fmt.Sprintf("%d", ff.Length))
	rw.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d", maxAge))
	rw.Write(ff.Content)
}
