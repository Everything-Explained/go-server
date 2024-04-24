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

func TestAssetRoute(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	rq := require.New(t)

	dir := t.TempDir()
	err := os.Mkdir(dir+"/mocks", 0o644)
	rq.NoError(err, "create mock dir")

	err = os.WriteFile(dir+"/mocks/mock.txt", []byte("test text"), 0o600)
	rq.NoError(err, "create mock file")

	r := router.NewRouter()
	HandleAssets(r, dir+"/mocks")

	rec := testutils.MockRequest(r.Handler, "GET", "/assets/mock.txt", nil)

	rq.Equal(http.StatusOK, rec.Code, "expect StatusOk")
	a.Equal("test text", rec.Body.String(), "returns expected body text")
}
