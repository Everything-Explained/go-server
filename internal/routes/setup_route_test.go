package routes

import (
	"net/http"
	"os"
	"testing"

	"github.com/Everything-Explained/go-server/internal/db"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Mock403 struct {
	Headers *map[string][]string
}

func TestSetupRoute(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	rq := require.New(t)

	dir := t.TempDir()

	u, err := db.NewUsers(dir)
	rq.NoError(err, "init users db")
	defer u.Close()

	err = os.WriteFile(dir+"/mock.txt", []byte("test text"), 0o600)
	rq.NoError(err, "write mock file")

	r := router.NewRouter()
	HandleSetup(r, dir+"/mock.txt", u)
	expBody := "Malformed Authorization\n"

	t.Run("detects bad authorization header", func(t *testing.T) {
		authTable := []string{
			"",
			" ",
			"gibberish",
			"Bearer",
			"Bearer ",
		}
		for _, entry := range authTable {
			var headers *map[string][]string
			if entry != "" {
				headers = &map[string][]string{
					"Authorization": {
						entry,
					},
				}
			}
			resp := testutils.MockRequest(r.Handler, "GET", "/setup", headers)
			assert.Equal(http.StatusForbidden, resp.Code, "forbid bad authentication")
			assert.Equal(expBody, resp.Body.String(), "return expected body")
		}
	})

	t.Run("detects unauthorized request", func(t *testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/setup", &map[string][]string{
			"Authorization": {"Bearer gibberish"},
		})

		assert.Equal(http.StatusUnauthorized, resp.Code, "expect unauthorized status")
		assert.Equal(
			"Authorization Expired or Missing\n",
			resp.Body.String(),
			"return expected body",
		)
	})

	t.Run("adds users", func(t *testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/setup", &map[string][]string{
			"Authorization": {"Bearer setup"},
		})

		assert.Equal(http.StatusOK, resp.Code, "expected status ok")

		id := resp.Header().Get("X-Evex-Id")
		assert.NotEmpty(id, "expected X-Evex-Id header to exist")

		isRed33med, err := u.GetState(id)
		assert.NoError(err, "user should exist")
		assert.False(isRed33med, "user should not have red33m access by default")
		assert.Equal("test text", resp.Body.String(), "returns expected body")
	})

	t.Run("detects authenticated user", func(t *testing.T) {
		id, _ := u.GetRandomUserId()
		resp := testutils.MockRequest(r.Handler, "GET", "/setup", &map[string][]string{
			"Authorization": {"Bearer " + id},
		})

		assert.Equal(http.StatusOK, resp.Code, "expected status ok")

		assert.Equal("test text", resp.Body.String(), "returns expected body")

		idHeader := resp.Header().Get("X-Evex-Id")
		assert.Empty(idHeader, "should not include ID header")

		red33mVal := resp.Header().Get("X-Evex-Red33m")
		assert.Equal("no", red33mVal, "should have red33m")
	})

	t.Run("detects redeem user", func(t *testing.T) {
		id, _ := u.GetRandomUserId()
		u.Update(id, true)
		resp := testutils.MockRequest(r.Handler, "GET", "/setup", &map[string][]string{
			"Authorization": {"Bearer " + id},
		})

		assert.Equal(http.StatusOK, resp.Code, "expected status ok")

		red33mVal := resp.Header().Get("X-Evex-Red33m")
		assert.Equal("yes", red33mVal, "should have red33m")
	})
}
