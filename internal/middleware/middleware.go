package middleware

import (
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter
	strStore map[string]string
	intStore map[string]int64
	status   int
}

type httpHandler = func(rw http.ResponseWriter, req *http.Request)

func (rw *ResponseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func NewHandler(handlers ...func(rw *ResponseWriter, req *http.Request)) httpHandler {
	return func(rw http.ResponseWriter, req *http.Request) {
		resWriter := &ResponseWriter{
			ResponseWriter: rw,
			strStore:       map[string]string{},
			intStore:       map[string]int64{},
		}
		for _, h := range handlers {
			h(resWriter, req)
		}
	}
}
