package router

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Everything-Explained/go-server/internal"
)

type Method string

const (
	ANY    Method = ""
	GET    Method = "GET "
	POST   Method = "POST "
	PUT    Method = "PUT "
	PATCH  Method = "PATCH "
	DELETE Method = "DELETE "
)

var reqBodyKey = &internal.ContextKey{Name: "body"}

func NewRouter() *Router {
	sx := http.NewServeMux()
	return &Router{
		Handler: sx,
	}
}

func AddSubRoute(path string, parentRouter *Router, childRouter *Router, mw ...Middleware) {
	if path == "/" {
		panic("sub-route cannot be the root route")
	}

	if path[:len(path)-1] == "/" {
		panic("sub-route cannot have trailing forward slash '/'")
	}

	if childRouter.mwCount > 0 && len(mw) > 0 {
		panic(
			"route-level middleware is not allowed with sub-route middleware; use one or the other",
		)
	}

	if len(mw) > 0 {
		parentRouter.createHandler(
			path+"/",
			ANY,
			http.StripPrefix(path, childRouter.Handler),
			mw...)
		return
	}

	parentRouter.createHandler(path+"/", ANY, http.StripPrefix(path, childRouter.Handler))
}

type Router struct {
	Handler *http.ServeMux
	mwCount int
}

/*
Any sets up a route that accepts all methods (GET, POST, etc...)
*/
func (r *Router) Any(route string, handler http.HandlerFunc, mw ...Middleware) {
	r.createHandler(route, ANY, handler, mw...)
}

func (r *Router) Get(route string, handler http.HandlerFunc, mw ...Middleware) {
	r.createHandler(route, GET, handler, mw...)
}

func (r *Router) Post(route string, handler http.HandlerFunc, mw ...Middleware) {
	r.createHandler(route, POST, handler, mw...)
}

func (r *Router) SetStaticRoute(
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
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		file := req.PathValue("file")
		err := FileServer.ServeMaxCache(folderPath+"/"+file, rw)
		if err != nil {
			// TODO  Log error
			panic(err)
		}
	}, mw...)
}

func (*Router) ListenAndServe(addr string, port int) error {
	s := http.Server{
		Addr:         fmt.Sprintf("%s:%d", addr, port),
		ReadTimeout:  8 * time.Second,
		WriteTimeout: 8 * time.Second,
	}
	return s.ListenAndServe()
}

func GetContextValue[T any](key any, r *http.Request) (T, error) {
	v, ok := r.Context().Value(key).(T)
	if !ok {
		return v, fmt.Errorf("could not find context key: %v", key)
	}
	return v, nil
}

func GetBody(r *http.Request) string {
	body, ok := r.Context().Value(reqBodyKey).(string)
	if !ok {
		panic("request is missing 'body' context")
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
		panic("all route paths must start with a forward slash: '/'")
	}

	if strings.Contains(path, " ") {
		panic("route paths cannot contain spaces")
	}

	var chain Middleware
	if len(mw) > 0 {
		chain = CreateMiddlewareChain(mw...)
		r.mwCount += len(mw)
	}

	route := fmt.Sprintf("%s%s", m, path)

	if chain != nil {
		r.Handler.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			reqWithBody := r.WithContext(
				context.WithValue(r.Context(), reqBodyKey, readBody(r)),
			)
			chain(handler).ServeHTTP(w, reqWithBody)
		})
		return
	}

	r.Handler.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		reqWithBody := r.WithContext(context.WithValue(r.Context(), reqBodyKey, readBody(r)))
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
