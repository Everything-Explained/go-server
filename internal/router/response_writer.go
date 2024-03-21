package router

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

/*
ResponseWriter is a custom interface that extends the functionality
of the existing http.ResponseWriter. This allows router middleware
to access the status before & after a response, as well as
maintain state through Context.
*/
type ResponseWriter interface {
	http.ResponseWriter
	WithValue(key any, val any)
	GetStatus() int
	GetContextValue(key any) any

	/*
	   GetBody gets a cached version of the http.Request.Body field.

	   🟠 Use this method to get the body instead of http.Request.Body,
	   as it will always be empty.
	*/
	GetBody() string
}

type ContextKey struct{ Name string }

var (
	respStatusKey = &ContextKey{"status"}
	respBodyKey   = &ContextKey{"body"}
)

func NewResponseWriter(rw http.ResponseWriter, req *http.Request) *responseWriter {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		// todo - log to server error log
		panic(err)
	}

	// Initialize status to 200 to protect against fall-through
	ctx := context.WithValue(context.Background(), respStatusKey, 200)
	ctx = context.WithValue(ctx, respBodyKey, string(body))

	return &responseWriter{
		ResponseWriter: rw,
		context:        ctx,
	}
}

func GetContextValue[T any](key any, rw ResponseWriter) (T, error) {
	var zero T
	if v := rw.GetContextValue(key); v == nil {
		return zero, fmt.Errorf("cannot find '%s' context", key)
	} else {
		return v.(T), nil
	}
}

type responseWriter struct {
	http.ResponseWriter
	context context.Context
}

func (rw *responseWriter) WithValue(key any, val any) {
	rw.context = context.WithValue(rw.context, key, val)
}

func (rw *responseWriter) GetContextValue(key any) any {
	return rw.context.Value(key)
}

func (rw *responseWriter) GetStatus() int {
	status, _ := GetContextValue[int](respStatusKey, rw)
	return status
}

func (rw *responseWriter) GetBody() string {
	body, _ := GetContextValue[string](respBodyKey, rw)
	return body
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.WithValue(respStatusKey, status)
	rw.ResponseWriter.WriteHeader(status)
}
