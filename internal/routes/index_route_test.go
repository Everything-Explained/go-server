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
	dir := t.TempDir()
	r := router.NewRouter()
	err := os.WriteFile(dir+"/mock.html", []byte("index text"), 0o600)
	require.NoError(t, err, "write mock file")

	HandleIndex(r, dir+"/mock.html")

	t.Run("panic if handling index without file", func(t *testing.T) {
		r := router.NewRouter()
		assert.PanicsWithValue(
			t,
			"index route needs a file path, not folder path",
			func() { HandleIndex(r, dir+"/mock") },
			"panic when passing an invalid index file",
		)
	})

	t.Run("returns index on root url request", func(t *testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/", nil)
		assert.Equal(t, http.StatusOK, resp.Code, "returns status ok")
		assert.Equal(t, "index text", resp.Body.String(), "return index file")
		assert.Equal(
			t,
			"text/html; charset=utf-8",
			resp.Result().Header.Get("Content-Type"),
			"should have html content type",
		)
	})

	t.Run("returns index on index.html request", func(t *testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/index.html", nil)
		assert.Equal(t, http.StatusOK, resp.Code, "return status ok")
		assert.Equal(t, "index text", resp.Body.String(), "return index contents")
	})

	t.Run("returns index on unknown uri request", func(t *testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/unknown/url", nil)
		assert.Equal(t, http.StatusOK, resp.Code, "return status ok")
		assert.Equal(t, "index text", resp.Body.String(), "return index contents")
	})

	t.Run("return 404 on file request", func(t *testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/unknown/file.ext", nil)
		assert.Equal(t, http.StatusNotFound, resp.Code, "return status 'not found'")
	})
}
