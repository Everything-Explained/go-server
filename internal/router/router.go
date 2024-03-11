package router

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Everything-Explained/go-server/internal/lib"
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
		// Enable or disable logging (off by default)
		CanLog bool
	}
)

func NewRouter() *Router {
	return &Router{}
}

type Router struct{}

// Get handles the GET method for the specified path and accepts
// middlewares, including the main handler for this route. The
// handlers are executed in the order they are declared.
//
// NOTE: Handlers execute one after the other; there is no way
// to pause or stop the chain. If you need to guard (stop)
// a route
func (r *Router) Get(path string, handlers ...HTTPFunc) {
	validatePath(path)
	route := fmt.Sprintf("GET %s", path)
	createHandler(route, handlers)
}

/*
AddStaticRoute sets up a route handler with the specified path, to serve
files from the specified folder, using FastFileServer.
*/
func (r *Router) AddStaticRoute(path string, folder string) {
	staticDir := fmt.Sprintf("C:\\Server\\evex-production\\client\\%s", folder)
	if _, err := os.Stat(staticDir); err != nil {
		if strings.Contains(err.Error(), "cannot find") {
			panic(fmt.Sprintf("static directory does not exist: %s", staticDir))
		}
		panic(err)
	}
	r.Get(
		fmt.Sprintf("%s/{file}", path),
		// NOTE  Maybe cache 404 requests in the future (micro-optimization)
		func(rw *http_interface.ResponseWriter, req *http.Request) {
			if !strings.Contains(req.URL.Path, ".") {
				rw.WriteHeader(404)
				return
			}
			file := req.PathValue("file")
			ff, err := lib.FastFileServer(fmt.Sprintf("%s\\%s", staticDir, file), "")
			if err != nil {
				if os.IsNotExist(err) {
					rw.WriteHeader(404)
					return
				}
				panic(err)
			}
			rw.Header().Add("Content-Type", ff.ContentType)
			rw.Header().Add("Content-Length", fmt.Sprintf("%d", ff.Length))
			rw.Write(ff.Content)
		},
	)
}

func (r *Router) Post(path string, handlers ...HTTPFunc) {
	validatePath(path)
	route := fmt.Sprintf("POST %s", path)
	createHandler(route, handlers)
}

func (r *Router) Patch(path string, handlers ...HTTPFunc) {
	validatePath(path)
	route := fmt.Sprintf("PATCH %s", path)
	createHandler(route, handlers)
}

func (r *Router) Delete(path string, handlers ...HTTPFunc) {
	validatePath(path)
	route := fmt.Sprintf("DELETE %s", path)
	createHandler(route, handlers)
}

func (r *Router) Listen(addr string, port int) {
	http.ListenAndServe(fmt.Sprintf("%s:%d", addr, port), nil)
}

/*
AddGetGuard guards the specified path string with a function that returns
a message and HTTP status code. A code >= to 400 results in the message
and status code being written to the response, skipping handler and
all middleware execution.

📝 Because middleware cannot be executed before the guard, the logging
middleware has been included, behind the flag: GuardData.CanLog

🔴 Panics if no handler is provided in GuardData
*/
func (r *Router) AddGetGuard(path string, guard GuardFunc, gd GuardData) {
	validatePath(path)
	pattern := fmt.Sprintf("GET %s", path)
	createGuardHandler(pattern, guard, gd)
}

func (r *Router) AddPostGuard(path string, guard GuardFunc, data GuardData) {
	validatePath(path)
	pattern := fmt.Sprintf("POST %s", path)
	createGuardHandler(pattern, guard, data)
}

func (r *Router) AddPatchGuard(path string, guard GuardFunc, data GuardData) {
	validatePath(path)
	pattern := fmt.Sprintf("PATCH %s", path)
	createGuardHandler(pattern, guard, data)
}

func (r *Router) AddDeleteGuard(path string, guard GuardFunc, data GuardData) {
	validatePath(path)
	pattern := fmt.Sprintf("DELETE %s", path)
	createGuardHandler(pattern, guard, data)
}

func createHandler(route string, mw Middleware) {
	if len(mw) == 0 {
		panic("route needs at least one handler function")
	}
	http.HandleFunc(route, func(rw http.ResponseWriter, req *http.Request) {
		customResWriter := http_interface.CreateResponseWriter(rw, req)
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

func createGuardHandler(pattern string, guard GuardFunc, gd GuardData) {
	if gd.Handler == nil {
		panic("the default handler for a route guard cannot be nil")
	}

	http.Handle(pattern, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		customResWriter := http_interface.CreateResponseWriter(rw, req)

		if gd.CanLog {
			middleware.LogHandler.IncomingReq(customResWriter, req)
			defer func() {
				middleware.LogHandler.OutgoingResp(customResWriter, req)
			}()
		}

		msg, status := guard(customResWriter, req)
		if status >= 400 {
			customResWriter.WriteHeader(status)
			fmt.Fprint(customResWriter, msg)
			return
		}

		if len(gd.PreMiddleware) > 0 {
			for _, f := range gd.PreMiddleware {
				f(customResWriter, req)
			}
		}

		gd.Handler(customResWriter, req)

		if len(gd.PostMiddleware) > 0 {
			for _, f := range gd.PostMiddleware {
				f(customResWriter, req)
			}
		}
	}))
}
