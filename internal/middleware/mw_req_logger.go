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
LogRequests returns a middleware that logs all requests which respond
with a status code greater than or equal to the provided statusCode.
*/
func LogRequests(statusCode int) router.Middleware {
	const logName = "requests"
	if !writers.Log.Exists(logName) {
		err := writers.NewLogWriter(logName)
		if err != nil {
			panic(err)
		}
	}

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
			if err != nil {
				panic(err)
			}

			agent := strings.Join(req.Header["User-Agent"], ",")
			country := strings.Join(req.Header[http.CanonicalHeaderKey("CF-IPCountry")], ",")
			now := time.Now().UnixMicro()

			respWriterWrapper := &responseWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			body := router.ReadBody(req)

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
