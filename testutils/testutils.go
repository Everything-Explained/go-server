package testutils

import (
	"net/http"
	"net/http/httptest"
)

func MockRequest(
	h http.Handler,
	method string,
	uri string,
	headers *map[string][]string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, uri, nil)
	w := httptest.NewRecorder()
	if headers != nil {
		req.Header = *headers
	}
	h.ServeHTTP(w, req)
	return w
}
