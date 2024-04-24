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
	assert := assert.New(t)
	dir := t.TempDir()
	r := router.NewRouter()
	err := os.WriteFile(dir+"/mock.html", []byte("index text"), 0o600)
	require.NoError(t, err, "write mock file")

	HandleIndex(r, dir+"/mock.html")

	t.Run("panic if handling index without file", func(*testing.T) {
		r := router.NewRouter()
		assert.PanicsWithValue(
			"index route needs a file path, not folder path",
			func() { HandleIndex(r, dir+"/mock") },
			"panic when passing an invalid index file",
		)
	})

	t.Run("returns index on root url request", func(*testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/", nil)
		assert.Equal(http.StatusOK, resp.Code, "returns status ok")
		assert.Equal("index text", resp.Body.String(), "return index file")
		assert.Equal(
			"text/html; charset=utf-8",
			resp.Result().Header.Get("Content-Type"),
			"should have html content type",
		)
	})

	t.Run("returns index on index.html request", func(*testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/index.html", nil)
		assert.Equal(http.StatusOK, resp.Code, "return status ok")
		assert.Equal("index text", resp.Body.String(), "return index contents")
	})

	t.Run("returns index on unknown uri request", func(*testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/unknown/url", nil)
		assert.Equal(http.StatusOK, resp.Code, "return status ok")
		assert.Equal("index text", resp.Body.String(), "return index contents")
	})

	t.Run("return 404 on file request", func(*testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/unknown/file.ext", nil)
		assert.Equal(http.StatusNotFound, resp.Code, "return status 'not found'")
	})
}
