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
	a := assert.New(t)
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

	t.Run("detects bad authorization header", func(*testing.T) {
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
			a.Equal(http.StatusForbidden, resp.Code, "forbid bad authentication")
			a.Equal(expBody, resp.Body.String(), "return expected body")
		}
	})

	t.Run("detects unauthorized request", func(*testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/setup", &map[string][]string{
			"Authorization": {"Bearer gibberish"},
		})

		a.Equal(http.StatusUnauthorized, resp.Code, "expect unauthorized status")
		a.Equal(
			"Authorization Expired or Missing\n",
			resp.Body.String(),
			"return expected body",
		)
	})

	t.Run("adds users", func(*testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/setup", &map[string][]string{
			"Authorization": {"Bearer setup"},
		})

		a.Equal(http.StatusOK, resp.Code, "expected status ok")

		id := resp.Header().Get("X-Evex-Id")
		a.NotEmpty(id, "expected X-Evex-Id header to exist")

		isRed33med, err := u.GetState(id)
		a.NoError(err, "user should exist")
		a.False(isRed33med, "user should not have red33m access by default")
		a.Equal("test text", resp.Body.String(), "returns expected body")
	})

	t.Run("detects authenticated user", func(*testing.T) {
		id, _ := u.GetRandomUserId()
		resp := testutils.MockRequest(r.Handler, "GET", "/setup", &map[string][]string{
			"Authorization": {"Bearer " + id},
		})

		a.Equal(http.StatusOK, resp.Code, "expected status ok")

		a.Equal("test text", resp.Body.String(), "returns expected body")

		idHeader := resp.Header().Get("X-Evex-Id")
		a.Empty(idHeader, "should not include ID header")

		red33mVal := resp.Header().Get("X-Evex-Red33m")
		a.Equal("no", red33mVal, "should have red33m")
	})

	t.Run("detects redeem user", func(*testing.T) {
		id, _ := u.GetRandomUserId()
		u.Update(id, true)
		resp := testutils.MockRequest(r.Handler, "GET", "/setup", &map[string][]string{
			"Authorization": {"Bearer " + id},
		})

		a.Equal(http.StatusOK, resp.Code, "expected status ok")

		red33mVal := resp.Header().Get("X-Evex-Red33m")
		a.Equal("yes", red33mVal, "should have red33m")
	})
}
