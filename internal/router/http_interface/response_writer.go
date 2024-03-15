package http_interface

import (
	"context"
	"io"
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter
	context context.Context
	status  int
	body    string
}

func (rw *ResponseWriter) GetContext() context.Context {
	return rw.context
}

func (rw *ResponseWriter) SetContext(ctx context.Context) {
	rw.context = ctx
}

func (rw *ResponseWriter) GetStatus() int {
	return rw.status
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	i, err := rw.ResponseWriter.Write(b)
	if err != nil {
		return i, err
	}
	rw.status = 200
	return i, nil
}

/*
GetBody gets the cached body of the http.Request that this ResponseWriter
is paired with.

🟠 Use this method to get the body instead of http.Request.Body,
as it will always be empty in the presence of this ResponseWriter.
*/
func (rw *ResponseWriter) GetBody() string {
	return string(rw.body)
}

func (rw *ResponseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func CreateResponseWriter(rw http.ResponseWriter, req *http.Request) *ResponseWriter {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		// todo - log to server error log
		panic(err)
	}
	return &ResponseWriter{
		ResponseWriter: rw,
		context:        context.Background(),
		body:           string(body),
	}
}
