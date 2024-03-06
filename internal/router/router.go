package router

import (
	"fmt"
	"net/http"
	"strings"
)

type (
	HTTPFunc   = func(rw *ResponseWriter, req *http.Request)
	Middleware = []HTTPFunc

	GuardFunc = func(rw *ResponseWriter, req *http.Request) (string, int)
	GuardData struct {
		// Middleware that is executed before the handler
		PreMiddleware []HTTPFunc
		// Middleware that is executed after the handler
		PostMiddleware []HTTPFunc
		// Function responsible for main route functionality
		Handler HTTPFunc
	}
)

const (
	maxIntStoreSize = 10
	maxStrStoreSize = 10
)

func NewRouter() *router {
	return &router{}
}

type router struct{}

// Get handles the GET method for the specified path and accepts
// middlewares, including the main handler for this route. The
// handlers are executed in the order they are declared.
//
// NOTE: Handlers execute one after the other; there is no way
// to pause or stop the chain. If you need to guard (stop)
// a route
func (r *router) Get(path string, handlers ...HTTPFunc) {
	validatePath(path)
	route := fmt.Sprintf("GET %s", path)
	createHandler(route, handlers)
}

func (r *router) Post(path string, handlers ...HTTPFunc) {
	validatePath(path)
	route := fmt.Sprintf("POST %s", path)
	createHandler(route, handlers)
}

func (r *router) Patch(path string, handlers ...HTTPFunc) {
	validatePath(path)
	route := fmt.Sprintf("PATCH %s", path)
	createHandler(route, handlers)
}

func (r *router) Delete(path string, handlers ...HTTPFunc) {
	validatePath(path)
	route := fmt.Sprintf("DELETE %s", path)
	createHandler(route, handlers)
}

func (r *router) Listen(addr string, port int) {
	http.ListenAndServe(fmt.Sprintf("%s:%d", addr, port), nil)
}

// AddGetGuard guards the specified path string with a function which returns
// a message and HTTP status code. A code >= to 400 results in the message
// and status code being written to the response, skipping handler and
// all middleware execution.
func (r *router) AddGetGuard(path string, guard GuardFunc, data GuardData) {
	validatePath(path)
	pattern := fmt.Sprintf("GET %s", path)
	createGuardHandler(pattern, guard, data)
}

func (r *router) AddPostGuard(path string, guard GuardFunc, data GuardData) {
	validatePath(path)
	pattern := fmt.Sprintf("POST %s", path)
	createGuardHandler(pattern, guard, data)
}

func (r *router) AddPatchGuard(path string, guard GuardFunc, data GuardData) {
	validatePath(path)
	pattern := fmt.Sprintf("PATCH %s", path)
	createGuardHandler(pattern, guard, data)
}

func (r *router) AddDeleteGuard(path string, guard GuardFunc, data GuardData) {
	validatePath(path)
	pattern := fmt.Sprintf("DELETE %s", path)
	createGuardHandler(pattern, guard, data)
}

func createHandler(route string, mw Middleware) {
	if len(mw) == 0 {
		panic("route needs at least one handler function")
	}
	http.HandleFunc(route, func(rw http.ResponseWriter, req *http.Request) {
		customResWriter := createResponseWriter(rw)
		for _, f := range mw {
			f(customResWriter, req)
		}
	})
}

func validatePath(path string) {
	if !strings.HasPrefix(path, "/") {
		panic("invalid path, all paths should start with a '/'")
	}
}

func createGuardHandler(pattern string, guard GuardFunc, data GuardData) {
	if data.Handler == nil {
		panic("the default handler for a route guard cannot be nil")
	}
	http.Handle(pattern, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		customResWriter := createResponseWriter(rw)
		msg, status := guard(customResWriter, req)
		if status >= 400 {
			rw.WriteHeader(status)
			fmt.Fprint(rw, msg)
			return
		}

		if len(data.PreMiddleware) > 0 {
			for _, f := range data.PreMiddleware {
				f(customResWriter, req)
			}
		}

		data.Handler(customResWriter, req)

		if len(data.PostMiddleware) > 0 {
			for _, f := range data.PostMiddleware {
				f(customResWriter, req)
			}
		}
	}))
}

func createResponseWriter(rw http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: rw,
		strStore:       make(map[string]string, maxStrStoreSize),
		intStore:       make(map[string]int64, maxIntStoreSize),
	}
}

type ResponseWriter struct {
	http.ResponseWriter
	strStore map[string]string
	intStore map[string]int64
	status   int
}

func (rw *ResponseWriter) StoreStr(id string, val string) string {
	rw.strStore[id] = val
	return val
}

func (rw *ResponseWriter) StoreInt(id string, val int64) int64 {
	rw.intStore[id] = val
	return val
}

func (rw *ResponseWriter) GetInt(id string) int64 {
	return rw.intStore[id]
}

func (rw *ResponseWriter) GetStr(id string) string {
	return rw.strStore[id]
}

func (rw *ResponseWriter) GetStatus() int {
	return rw.status
}

func (rw *ResponseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
