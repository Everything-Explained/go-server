package http_interface

import "net/http"

const (
	maxIntStoreSize = 10
	maxStrStoreSize = 10
)

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

func CreateResponseWriter(rw http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: rw,
		strStore:       make(map[string]string, maxStrStoreSize),
		intStore:       make(map[string]int64, maxIntStoreSize),
	}
}
