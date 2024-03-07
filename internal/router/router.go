package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Everything-Explained/go-server/internal/router/http_interface"
	"github.com/Everything-Explained/go-server/internal/router/middleware"
)

type (
	HTTPFunc   = func(rw *http_interface.ResponseWriter, req *http.Request)
	Middleware = []HTTPFunc

	GuardFunc = func(rw *http_interface.ResponseWriter, req *http.Request) (string, int)
	GuardData struct {
		// Middleware that is executed before the handler
		PreMiddleware []HTTPFunc
		// Middleware that is executed after the handler
		PostMiddleware []HTTPFunc
		// Function responsible for main route functionality
		Handler HTTPFunc
	}
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
		customResWriter := http_interface.CreateResponseWriter(rw)
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
		customResWriter := http_interface.CreateResponseWriter(rw)
		log := middleware.LogHandler
		log.IncomingReq(customResWriter, req)
		msg, status := guard(customResWriter, req)
		if status >= 400 {
			rw.WriteHeader(status)
			fmt.Fprint(rw, msg)
			log.OutgoingResp(customResWriter, req)
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
