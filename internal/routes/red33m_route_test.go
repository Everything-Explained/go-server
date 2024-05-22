package routes

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/Everything-Explained/go-server/configs"
	"github.com/Everything-Explained/go-server/internal/db"
	"github.com/Everything-Explained/go-server/internal/middleware"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedeemRoute(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	rq := require.New(t)
	tmpDir := t.TempDir()

	cfg := mockConfig(t, tmpDir, "../../configs")

	u, err := db.NewUsers(tmpDir)
	rq.NoError(err, "mock users database")
	defer u.Close()

	uid, err := u.Add(false)
	rq.NoError(err, "add test user to database")

	r := router.NewRouter()
	HandleRed33m(r, u, cfg.Red33mPassword, middleware.AuthGuard(u))

	t.Run("detect missing body", func(*testing.T) {
		resp := testutils.MockRequest(r.Handler, "POST", "/red33m", nil, &map[string][]string{
			"Authorization": {"Bearer " + uid},
		})
		a.Equal(http.StatusBadRequest, resp.Code, "expect BadRequest status")
		a.Equal("Missing Body\n", resp.Body.String(), "expect missing body message")
	})

	t.Run("detect invalid password", func(*testing.T) {
		resp := testutils.MockRequest(
			r.Handler,
			"POST",
			"/red33m",
			strings.NewReader("invalidpass"),
			&map[string][]string{
				"Authorization": {"Bearer " + uid},
			},
		)
		a.Equal(http.StatusUnauthorized, resp.Code, "expect Unauthorized status")
		a.Equal("Invalid Password\n", resp.Body.String(), "expect body message")
	})

	t.Run("detect already logged in", func(*testing.T) {
		uid, err := u.Add(true)
		if err != nil {
			panic(err)
		}
		resp := testutils.MockRequest(
			r.Handler,
			"POST",
			"/red33m",
			strings.NewReader(""),
			&map[string][]string{
				"Authorization": {"Bearer " + uid},
			},
		)
		a.Equal(http.StatusBadRequest, resp.Code, "expect BadRequest status")
		a.Equal("Already Logged In\n", resp.Body.String(), "expect already logged in")
	})

	t.Run("successful login", func(*testing.T) {
		uid, err := u.Add(false)
		uState, err := u.GetState(uid)
		rq.NoError(err, "get mocked user state")
		rq.False(uState, "expect false red33m state")

		resp := testutils.MockRequest(
			r.Handler,
			"POST",
			"/red33m",
			strings.NewReader("D4DDY"),
			&map[string][]string{
				"Authorization": {"Bearer " + uid},
			},
		)

		a.Equal(http.StatusOK, resp.Code, "expect Ok status")
		a.Equal("Login Successful\n", resp.Body.String(), "expect login success")

		uState, err = u.GetState(uid)
		rq.NoError(err, "get mocked user")
		a.True(uState, "expect true red33m state")
	})
}

func mockConfig(t *testing.T, tmpDir string, cfgDir string) configs.ConfigData {
	rq := require.New(t)

	envFile, err := os.Open(cfgDir + "/.env.dev")
	rq.NoError(err, "get env file")

	dstEnvFile, err := os.Create(tmpDir + "/.env.dev")
	rq.NoError(err, "mock env file")

	_, err = io.Copy(dstEnvFile, envFile)
	rq.NoError(err, "copy env file data to mock file")
	envFile.Close()
	dstEnvFile.Close()

	cfgFile, err := os.Open(cfgDir + "/dev.config.yml")
	rq.NoError(err, "get config file")

	dstCfgFile, err := os.Create(tmpDir + "/dev.config.yml")
	rq.NoError(err, "mock config file")

	_, err = io.Copy(dstCfgFile, cfgFile)
	rq.NoError(err, "copy config file data to mock file")
	cfgFile.Close()
	dstCfgFile.Close()

	cfg, err := configs.GetConfig(tmpDir)
	rq.NoError(err, "get mock config data")
	return cfg
}
