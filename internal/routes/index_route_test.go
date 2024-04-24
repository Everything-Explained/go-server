package routes

import (
	"net/http"
	"os"
	"testing"

	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexRoute(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	dir := t.TempDir()
	r := router.NewRouter()
	err := os.WriteFile(dir+"/mock.html", []byte("index text"), 0o600)
	require.NoError(t, err, "write mock file")

	HandleIndex(r, dir+"/mock.html")

	t.Run("panic if handling index without file", func(*testing.T) {
		r := router.NewRouter()
		a.PanicsWithValue(
			"index route needs a file path, not folder path",
			func() { HandleIndex(r, dir+"/mock") },
			"panic when passing an invalid index file",
		)
	})

	t.Run("returns index on root url request", func(*testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/", nil)
		a.Equal(http.StatusOK, resp.Code, "returns status ok")
		a.Equal("index text", resp.Body.String(), "return index file")
		a.Equal(
			"text/html; charset=utf-8",
			resp.Result().Header.Get("Content-Type"),
			"should have html content type",
		)
	})

	t.Run("returns index on index.html request", func(*testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/index.html", nil)
		a.Equal(http.StatusOK, resp.Code, "return status ok")
		a.Equal("index text", resp.Body.String(), "return index contents")
	})

	t.Run("returns index on unknown uri request", func(*testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/unknown/url", nil)
		a.Equal(http.StatusOK, resp.Code, "return status ok")
		a.Equal("index text", resp.Body.String(), "return index contents")
	})

	t.Run("return 404 on file request", func(*testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/unknown/file.ext", nil)
		a.Equal(http.StatusNotFound, resp.Code, "return status 'not found'")
	})
}
