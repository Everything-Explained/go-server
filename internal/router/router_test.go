package router

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/Everything-Explained/go-server/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	a := assert.New(t)

	t.Run("route paths require a prefix", func(*testing.T) {
		r := NewRouter()
		a.Panics(
			func() {
				r.Get("", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
				})
			},
			"all route paths must start with a forward slash: '/'",
		)
	})

	t.Run("route should not contain spaces", func(*testing.T) {
		r := NewRouter()
		a.Panics(
			func() {
				r.Get("/foo bar", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
				})
			},
			"route paths cannot contain spaces",
		)
	})

	t.Run("validate static routes", func(*testing.T) {
		r := NewRouter()

		a.PanicsWithValue(
			"static directory does not exist: /foo/bar",
			func() {
				r.SetStaticRoute("/foo/bar", "/foo/bar")
			},
		)

		a.PanicsWithValue(
			"you provided a file path '/foo/bar.ext' instead of a folder path",
			func() {
				r.SetStaticRoute("/foo/bar", "/foo/bar.ext")
			},
		)
	})

	t.Run("static route should verify file exists", func(*testing.T) {
		r := NewRouter()
		tmpDir := t.TempDir()
		dir := tmpDir + "/foo"

		err := os.Mkdir(dir, 0o755)
		a.NoError(err, "create mock directory")

		a.NotPanics(func() {
			r.SetStaticRoute("/foo", dir)
		}, "set static route")

		resp := testutils.MockRequest(r.Handler, "GET", "/foo/mock", nil, nil)
		a.Equal(http.StatusNotFound, resp.Code, "expect status not found")
	})

	t.Run("should validate sub-routes", func(*testing.T) {
		r := NewRouter()
		a.PanicsWithValue("sub-route cannot be the root route", func() {
			AddSubRoute("/", r, r, nil)
		}, "should panic with root route")

		a.PanicsWithValue("sub-route cannot have trailing forward slash '/'", func() {
			AddSubRoute("/foo/", r, r, nil)
		}, "should panic with trailing slash")

		// Setup initial request with middleware
		r.Any("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		}, func(h http.Handler) http.Handler {
			return h
		})

		a.PanicsWithValue(
			"route-level middleware is not allowed with sub-route middleware; use one or the other",
			func() {
				AddSubRoute("/bar", r, r, func(h http.Handler) http.Handler {
					return h
				})
			},
			"should panic with middleware",
		)
	})

	t.Run("ReadBody should read without replacing it", func(*testing.T) {
		r := NewRouter()
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			body := ReadBody(r)
			a.Equal("hello world", string(body), "should read body and reset it")
			moreBody, err := io.ReadAll(r.Body)
			a.NoError(err, "should read body as if it was not replaced")
			w.Write(moreBody)
		})

		resp := testutils.MockRequest(
			r.Handler,
			"GET",
			"/",
			strings.NewReader("hello world"),
			nil,
		)

		a.Equal(resp.Body.String(), "hello world", "should read body")
	})
}
