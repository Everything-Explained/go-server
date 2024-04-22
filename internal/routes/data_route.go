package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal/middleware"
	"github.com/Everything-Explained/go-server/internal/router"
)

/*
HandleData serves as the root route for all meaningful content
on the site (literature, videos, etc...)

🟠 Requires the auth guard middleware.
*/
func HandleData(r *router.Router, dir string, mw ...router.Middleware) {
	r.Get(
		"/data/{content}/{visibility}",
		getSummaryDataHandler(dir),
		mw...,
	)

	r.Get(
		"/data/{content}/{visibility}/{file}",
		getMDHTMLHandler(dir),
		mw...,
	)
}

func getSummaryDataHandler(dataPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agData := middleware.GetAuthGuardData(r)
		content := r.PathValue("content")
		visibility := r.PathValue("visibility")

		if strings.Contains(visibility, "red33m") {
			if !agData.IsRed33med {
				http.Error(w, "Not Authorized", http.StatusUnauthorized)
				return
			}
		}

		// Only supports non-file requests
		if strings.Contains(visibility, ".") {
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

func getMDHTMLHandler(dataPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agData := middleware.GetAuthGuardData(r)
		content := r.PathValue("content")
		visibility := r.PathValue("visibility")
		file := r.PathValue("file")

		if strings.Contains(visibility, "red33m") {
			if !agData.IsRed33med {
				http.Error(w, "Not Authorized", http.StatusUnauthorized)
				return
			}
		}

		// Only supports MDHTML files
		if !strings.HasSuffix(file, ".mdhtml") {
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
