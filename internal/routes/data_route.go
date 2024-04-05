package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/middleware"
	"github.com/Everything-Explained/go-server/internal/router"
)

/*
HandleData serves as the root route for all meaningful content
on the site (literature, videos, etc...)

🟠 Requires the auth guard middleware.
*/
func HandleData(r *router.Router, mw ...router.Middleware) {
	r.Get(
		"/data/{content}/{visibility}",
		getSummaryDataHandler(),
		mw...,
	)

	r.Get(
		"/data/{content}/{visibility}/{file}",
		getMDHTMLHandler(),
		mw...,
	)
}

func getSummaryDataHandler() http.HandlerFunc {
	dataPath := configs.GetConfig().DataPath
	return func(w http.ResponseWriter, r *http.Request) {
		agData := middleware.GetAuthGuardData(r)
		content := r.PathValue("content")
		visibility := r.PathValue("visibility")
		notRed33med := strings.Contains(visibility, "red33m") && !agData.IsRed33med

		// Only supports non-file requests
		if strings.Contains(visibility, ".") || notRed33med {
			http.Error(w, "File Not Found", http.StatusNotFound)
			return
		}

		filePath := fmt.Sprintf("%s/%s/%s/%s.json", dataPath, content, visibility, visibility)
		err := router.FileServer.ServeFile(filePath, w, r, true)
		if err != nil {
			panic(err)
		}
	}
}

func getMDHTMLHandler() http.HandlerFunc {
	dataPath := configs.GetConfig().DataPath
	return func(w http.ResponseWriter, r *http.Request) {
		agData := middleware.GetAuthGuardData(r)
		content := r.PathValue("content")
		visibility := r.PathValue("visibility")
		file := r.PathValue("file")
		notRed33med := strings.Contains(visibility, "red33m") && !agData.IsRed33med

		// Only supports MDHTML files
		if !strings.HasSuffix(file, ".mdhtml") || notRed33med {
			http.Error(w, "File Not Found", http.StatusNotFound)
			return
		}

		filePath := fmt.Sprintf(
			"%s/%s/%s/%s",
			dataPath,
			content,
			visibility,
			file,
		)

		err := router.FileServer.ServeFile(filePath, w, r, true)
		if err != nil {
			panic(err)
		}
	}
}
