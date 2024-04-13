package testutils

import (
	"fmt"
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

// PrintErrorS returns a string formatted for simple test errors
func PrintErrorS(want any, got any) string {
	return fmt.Sprintf("\n\t(WANT= %v) \n\t(GOT= %v)", want, got)
}

/*
PrintErrorD returns a string formatted for descriptive test errors,
allowing for an expectation to describe what should be happening.
*/
func PrintErrorD(expected string, want any, got any) string {
	return fmt.Sprintf("\n\t(EXPECTED= %s) %s", expected, PrintErrorS(want, got))
}
