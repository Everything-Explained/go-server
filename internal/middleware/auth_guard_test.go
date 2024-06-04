package middleware

import (
	"net/http"
	"testing"

	"github.com/Everything-Explained/go-server/internal/db"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthGuard(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	rq := require.New(t)

	tmpDir := t.TempDir()

	u, err := db.NewUsers(tmpDir)
	rq.NoError(err, "should initialize new users")

	defer u.Close()
	u.Add(false)
	userID, err := u.GetRandomUserId()
	rq.NoError(err, "should get random user id")

	r := router.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}, AuthGuard(u))

	t.Run("halts req chain on bad authorization header", func(*testing.T) {
		req := testutils.MockRequest(
			r.Handler,
			"GET",
			"/",
			nil,
			nil, // test no header
		)

		a.Equal(http.StatusUnauthorized, req.Code, "should return unauthorized status")
		a.Equal("Malformed Authorization\n", req.Body.String(), "should return reason")

		req = testutils.MockRequest(
			r.Handler,
			"GET",
			"/",
			nil,
			&map[string][]string{
				"Authorization": {"Bearer "}, // test invalid header
			},
		)

		a.Equal(http.StatusUnauthorized, req.Code, "should return unauthorized status")
		a.Equal("Malformed Authorization\n", req.Body.String(), "should return reason")
	})

	t.Run("halts req chain on unauthorized users", func(*testing.T) {
		req := testutils.MockRequest(
			r.Handler,
			"GET",
			"/",
			nil,
			&map[string][]string{
				"Authorization": {"Bearer testuser"},
			},
		)

		a.Equal(http.StatusForbidden, req.Code, "should return forbidden status")
		a.Equal("Bad User\n", req.Body.String(), "should return reason")
	})

	t.Run("resumes req chain on valid users", func(*testing.T) {
		req := testutils.MockRequest(
			r.Handler,
			"GET",
			"/",
			nil,
			&map[string][]string{
				"Authorization": {"Bearer " + userID},
			},
		)

		a.Equal(http.StatusOK, req.Code, "should return okay status")
	})
}
