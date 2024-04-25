package testutils

import (
	"io"
	"net/http"
	"net/http/httptest"
)

func MockRequest(
	h http.Handler,
	method string,
	uri string,
	body io.Reader,
	headers *map[string][]string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, uri, body)
	w := httptest.NewRecorder()
	if headers != nil {
		req.Header = *headers
	}
	h.ServeHTTP(w, req)
	return w
}
