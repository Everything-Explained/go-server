package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/internal/writers"
)

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWrapper) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

/*
LogRequests creates a log at the specified directory and returns a logByStatus(),
to filter by minimum status code. If the status is set to 400, then no status
below 400 will be logged. When you're done with the log, you can use the
returned closeLog().
*/
func LogRequests(dir string) (closeLog func(), logByStatus func(status int) router.Middleware) {
	const logName = "requests"
	err := writers.CreateLog(logName, dir)
	if err != nil {
		panic(err)
	}

	closeLog = func() {
		writers.Log.Close(logName)
	}

	logByStatus = func(statusCode int) router.Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				query := ""
				if req.URL.RawQuery != "" {
					query = req.URL.RawQuery
				}

				host := req.Host
				if host == "" {
					host = req.RemoteAddr
				}

				url, err := req.URL.Parse(req.URL.RequestURI())
				// TODO  Log it as server error
				if err != nil {
					panic(err)
				}

				agent := strings.Join(req.Header["User-Agent"], ",")
				country := strings.Join(
					req.Header[http.CanonicalHeaderKey("CF-IPCountry")],
					",",
				)
				now := time.Now().UnixMicro()

				respWriterWrapper := &responseWrapper{
					ResponseWriter: w,
					statusCode:     http.StatusOK,
				}
				body := router.ReadBody(req)

				// Complete request ops before we log
				next.ServeHTTP(respWriterWrapper, req)

				if respWriterWrapper.statusCode < statusCode {
					return
				}

				reqSpeed := fmt.Sprintf("%dµs", time.Now().UnixMicro()-now)

				writers.Log.Info(
					logName,
					agent,
					country,
					req.Method,
					host,
					url.Path,
					query,
					body,
					respWriterWrapper.statusCode,
					reqSpeed,
				)
			})
		}
	}

	return
}
