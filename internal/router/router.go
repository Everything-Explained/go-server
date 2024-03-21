package router

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

type (
	HandlerFunc = func(rw ResponseWriter, req *http.Request)
	GuardFunc   = func(rw ResponseWriter, req *http.Request) (string, int)
)

type RouteData struct {
	// Middleware that is executed before the handler
	PreMiddleware []HandlerFunc
	// Middleware that is executed after the handler
	PostMiddleware []HandlerFunc
	// Function responsible for main route functionality
	Handler HandlerFunc
}

func NewRouter() *Router {
	return &Router{}
}

type Router struct{}

/*
Get handles the GET method for the specified path and accepts
middlewares, including the main handler for this route. The
handlers are executed in the order they are declared.

🔴 Panics if there are no handlers provided.

🟠 Handlers execute one after the other; there is no way
to pause or stop the chain of their execution. Use a
guard route if you need to protect a specific handler.
*/
func (r *Router) Get(path string, handlers ...HandlerFunc) {
	createHandler(path, "GET", handlers)
}

func (r *Router) Post(path string, handlers ...HandlerFunc) {
	createHandler(path, "POST", handlers)
}

func (r *Router) Listen(addr string, port int) {
	// TODO  Return error
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
func (r *Router) AddGetGuard(path string, guard GuardFunc, rd RouteData) {
	createGuardHandler(path, "GET", guard, rd)
}

func (r *Router) AddPostGuard(path string, guard GuardFunc, data RouteData) {
	createGuardHandler(path, "POST", guard, data)
}

/*
AddStaticRoute serves files from the specified folder path.

📝 Pre/Post Middleware is always executed, even if the file is
404 not found.

🟡 Does NOT serve files from sub-folders.
*/
func (r *Router) AddStaticRoute(route string, folderPath string, rd RouteData) {
	if rd.Handler != nil {
		panic("static route ignores handler; use middleware only")
	}

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

	r.Get(fmt.Sprintf("%s/{file}", route), func(rw ResponseWriter, req *http.Request) {
		execHandlers(rw, req, rd.PreMiddleware...)
		defer execHandlers(rw, req, rd.PostMiddleware...)

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
	})
}

func createHandler(path string, method string, handlers []HandlerFunc) {
	if !strings.HasPrefix(path, "/") {
		panic("invalid path, all paths should start with a: /")
	}

	if len(handlers) == 0 {
		panic("route needs at least one handler function")
	}

	route := fmt.Sprintf("%s %s", method, path)
	http.HandleFunc(route, func(rw http.ResponseWriter, req *http.Request) {
		execHandlers(NewResponseWriter(rw, req), req, handlers...)
	})
}

func createGuardHandler(path string, method string, guard GuardFunc, gd RouteData) {
	if !strings.HasPrefix(path, "/") {
		panic("invalid path, all paths should start with a: /")
	}

	if gd.Handler == nil {
		panic("the default handler for a route guard cannot be nil")
	}

	pattern := fmt.Sprintf("%s %s", method, path)
	http.Handle(pattern, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		customResWriter := NewResponseWriter(rw, req)

		execHandlers(customResWriter, req, gd.PreMiddleware...)
		defer func() {
			execHandlers(customResWriter, req, gd.PostMiddleware...)
		}()

		msg, status := guard(customResWriter, req)
		if status >= 400 {
			customResWriter.WriteHeader(status)
			fmt.Fprint(customResWriter, msg)
			return
		}

		gd.Handler(customResWriter, req)
	}))
}

func execHandlers(rw ResponseWriter, req *http.Request, handlers ...HandlerFunc) {
	if len(handlers) > 0 {
		for _, f := range handlers {
			f(rw, req)
		}
	}
}
