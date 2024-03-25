package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/routes/guards"
)

func HandleData(r *router.Router) {
	ag := guards.GetAuthGuard()
	r.GetWithGuard("/data/{content}/{visibility}", ag.HandlerFunc, router.RouteData{
		HandlerFunc: getDataSummaryHandler(ag),
	})
	r.GetWithGuard("/data/{content}/{visibility}/{file}", ag.HandlerFunc, router.RouteData{
		HandlerFunc: getMDHTMLHandler(ag),
	})
}

func getMDHTMLHandler(ag guards.AuthGuard) router.HandlerFunc {
	dataPath := configs.GetConfig().DataPath

	return func(rw router.ResponseWriter, req *http.Request) {
		ctxVal, err := ag.GetContextValue(rw)
		if err != nil {
			panic(err)
		}
		content := req.PathValue("content")
		visibility := req.PathValue("visibility")
		file := req.PathValue("file")
		notRed33med := strings.Contains(visibility, "red33m") && !ctxVal.IsRed33med

		// Only supports MDHTML files; all other requests are abnormal
		if !strings.HasSuffix(file, ".mdhtml") || notRed33med {
			rw.WriteHeader(403)
			fmt.Fprintf(rw, "suspicious activity")
			return
		}

		filePath := fmt.Sprintf(
			"%s/%s/%s/%s",
			dataPath,
			content,
			visibility,
			file,
		)
		err = router.FileServer.ServeMaxCache(filePath, rw, req)
		if err != nil {
			panic(err)
		}
	}
}

func getDataSummaryHandler(ag guards.AuthGuard) router.HandlerFunc {
	dataPath := configs.GetConfig().DataPath

	return func(rw router.ResponseWriter, req *http.Request) {
		ctxVal, err := ag.GetContextValue(rw)
		if err != nil {
			panic(err)
		}
		content := req.PathValue("content")
		visibility := req.PathValue("visibility")
		notRed33med := strings.Contains(visibility, "red33m") && !ctxVal.IsRed33med

		// Only supports non-file requests; all other requests are abnormal
		if strings.Contains(visibility, ".") || notRed33med {
			rw.WriteHeader(403)
			fmt.Fprintf(rw, "suspicious activity detected")
			return
		}

		filePath := fmt.Sprintf("%s/%s/%s/%s.json", dataPath, content, visibility, visibility)
		err = router.FileServer.ServeMaxCache(filePath, rw, req)
		if err != nil {
			panic(err)
		}
	}
}
