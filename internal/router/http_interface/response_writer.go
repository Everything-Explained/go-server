package http_interface

import (
	"io"
	"net/http"
)

const (
	maxIntStoreSize  = 10
	maxStrStoreSize  = 10
	maxBoolStoreSize = 10
)

type ResponseWriter struct {
	http.ResponseWriter
	strStore  map[string]string
	intStore  map[string]int64
	boolStore map[string]bool
	status    int
	body      string
}

func (rw *ResponseWriter) StoreStr(id string, val string) string {
	rw.strStore[id] = val
	return val
}

func (rw *ResponseWriter) StoreInt(id string, val int64) int64 {
	rw.intStore[id] = val
	return val
}

func (rw *ResponseWriter) StoreBool(id string, val bool) bool {
	rw.boolStore[id] = val
	return val
}

func (rw *ResponseWriter) GetInt(id string) int64 {
	return rw.intStore[id]
}

func (rw *ResponseWriter) GetStr(id string) string {
	return rw.strStore[id]
}

func (rw *ResponseWriter) GetBool(id string) bool {
	return rw.boolStore[id]
}

func (rw *ResponseWriter) GetStatus() int {
	return rw.status
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
		strStore:       make(map[string]string, maxStrStoreSize),
		intStore:       make(map[string]int64, maxIntStoreSize),
		boolStore:      make(map[string]bool, maxBoolStoreSize),
		body:           string(body),
	}
}
