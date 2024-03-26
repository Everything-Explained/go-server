package router

import "net/http"

type Middleware func(http.Handler) http.Handler

func CreateMiddlewareChain(mw ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			next = mw[i](next)
		}
		return next
	}
}
