package router

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Everything-Explained/go-server/internal"
)

type Method string

const (
	GET    Method = "GET"
	POST   Method = "POST"
	PUT    Method = "PUT"
	PATCH  Method = "PATCH"
	DELETE Method = "DELETE"
)

var ReqBodyKey = &internal.ContextKey{Name: "body"}

func NewRouter() *Router {
	sx := http.NewServeMux()
	return &Router{
		Handler: sx,
	}
}

type Router struct {
	Handler *http.ServeMux
}

func GetContextValue[T any](key any, r *http.Request) (T, error) {
	v, ok := r.Context().Value(key).(T)
	if !ok {
		return v, fmt.Errorf("could not find context key: %v", key)
	}
	return v, nil
}

func (r *Router) Get(route string, handler http.HandlerFunc, mw ...Middleware) {
	r.createHandler(route, GET, handler, mw...)
}

func (r *Router) Post(route string, handler http.HandlerFunc, mw ...Middleware) {
	r.createHandler(route, POST, handler, mw...)
}

func (r *Router) GetStatic(
	route string,
	folderPath string,
	mw ...Middleware,
) {
	if strings.Contains(folderPath, ".") {
		panic(
			fmt.Sprintf("you provided a file path '%s' instead of a folder path.", folderPath),
		)
	}

	if _, err := os.Stat(folderPath); err != nil {
		if os.IsNotExist(err) {
			panic(fmt.Sprintf("static directory does not exist: %s", folderPath))
		}
		panic(err)
	}

	r.Get(fmt.Sprintf("%s/{file}", route), func(rw http.ResponseWriter, req *http.Request) {
		if !strings.Contains(req.URL.Path, ".") {
			rw.WriteHeader(404)
			return
		}

		file := req.PathValue("file")
		err := FileServer.ServeMaxCache(
			folderPath+"/"+file,
			rw,
			req,
		)
		if err != nil {
			// TODO  Log error
			panic(err)
		}
	}, mw...)
}

func GetBody(r *http.Request) string {
	body, ok := r.Context().Value(ReqBodyKey).(string)
	if !ok {
		panic("missing body context")
	}
	return body
}

func (r *Router) createHandler(
	path string,
	m Method,
	handler http.Handler,
	mw ...Middleware,
) {
	if !strings.HasPrefix(path, "/") {
		panic("invalid path, all paths should start with a: /")
	}

	var chain Middleware
	if len(mw) > 0 {
		chain = CreateMiddlewareChain(mw...)
	}

	route := fmt.Sprintf("%s %s", m, path)

	if chain != nil {
		r.Handler.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			reqWithBody := r.WithContext(
				context.WithValue(r.Context(), ReqBodyKey, readBody(r)),
			)
			chain(handler).ServeHTTP(w, reqWithBody)
		})
		return
	}

	r.Handler.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		reqWithBody := r.WithContext(context.WithValue(r.Context(), ReqBodyKey, readBody(r)))
		handler.ServeHTTP(w, reqWithBody)
	})
}

func readBody(r *http.Request) string {
	if r.Body == nil {
		return ""
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(body))
}
