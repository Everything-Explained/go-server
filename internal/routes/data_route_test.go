package routes

import (
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/Everything-Explained/go-server/internal/db"
	"github.com/Everything-Explained/go-server/internal/middleware"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataRoute(t *testing.T) {
	a := assert.New(t)
	rq := require.New(t)

	t.Parallel()
	tmpDir := t.TempDir()
	_ = os.Mkdir(tmpDir+"/blog", fs.FileMode(os.O_WRONLY))
	_ = os.Mkdir(tmpDir+"/blog/public", fs.FileMode(os.O_WRONLY))
	_ = os.Mkdir(tmpDir+"/blog/red33m", fs.FileMode(os.O_WRONLY))

	u, err := db.NewUsers(tmpDir)
	rq.NoError(err, "should initialize new users")

	defer u.Close()
	u.Add(false)
	userID, err := u.GetRandomUserId()
	rq.NoError(err, "should get random user id")

	t.Run("panic on missing auth middleware", func(*testing.T) {
		r := router.NewRouter()
		HandleData(r, tmpDir)
		a.PanicsWithValue(
			"missing auth guard data; did you forget to add the auth guard middleware?",
			func() {
				testutils.MockRequest(r.Handler, "GET", "/data/blog/public", nil)
			},
			"should panic because of missing auth guard",
		)
	})

	t.Run("passes summary uri spec", func(*testing.T) {
		r := router.NewRouter()
		_ = os.WriteFile(tmpDir+"/blog/public/public.json", []byte("test text"), 0o600)
		HandleData(r, tmpDir, middleware.AuthGuard(u))

		rec := testutils.MockRequest(
			r.Handler,
			"GET",
			"/data/blog/public",
			&map[string][]string{
				"Authorization": {"Bearer " + userID},
			},
		)

		a.Equal(200, rec.Code, "should be StatusOk")
		a.Equal(rec.Body.String(), "test text", "correct body text")

		contentType := rec.Header().Get("Content-Type")
		wantedType := "application/json; charset=utf-8"
		a.Equal(contentType, wantedType, "should contain JSON content type")

		cacheControl := rec.Header().Get("Cache-Control")
		rq.Contains(cacheControl, "max-age=", "should have max-age cache control")

		// 3 months
		minMaxAge := 60 * 60 * 24 * 30 * 3
		age, err := strconv.Atoi(strings.Split(cacheControl, "max-age=")[1])
		rq.NoError(err, "convert max-age to integer")

		a.LessOrEqual(minMaxAge, age, "minimum max-age >= 3 months")
		a.NotEmpty(rec.Header().Get("Last-Modified"), "should have Last-Modified header")
	})

	t.Run("summary uri returns 404 when files requested", func(*testing.T) {
		r := router.NewRouter()
		HandleData(r, tmpDir, middleware.AuthGuard(u))

		rec := testutils.MockRequest(
			r.Handler,
			"GET",
			"/data/blog/public/public.json",
			&map[string][]string{
				"Authorization": {"Bearer " + userID},
			},
		)

		a.Equal(rec.Code, http.StatusNotFound)
	})

	t.Run("passes mdhtml uri spec", func(*testing.T) {
		r := router.NewRouter()
		HandleData(r, tmpDir, middleware.AuthGuard(u))
		_ = os.WriteFile(tmpDir+"/blog/public/1234567890.mdhtml", []byte("i am mdhtml"), 0o600)

		rec := testutils.MockRequest(
			r.Handler,
			"GET",
			"/data/blog/public/1234567890.mdhtml",
			&map[string][]string{
				"Authorization": {"Bearer " + userID},
			},
		)

		a.Equal(rec.Code, http.StatusOK, "status code")

		a.Equal(rec.Body.String(), "i am mdhtml", "expected body text")

		contentType := rec.Header().Get("Content-Type")
		wantedType := "text/html; charset=utf-8"
		a.Equal(contentType, wantedType, "should have expected content type")

		cacheControl := rec.Header().Get("Cache-Control")
		rq.Contains(cacheControl, "max-age=", "should have max-age cache control")

		// 3 months
		minMaxAge := 60 * 60 * 24 * 30 * 3
		age, err := strconv.Atoi(strings.Split(cacheControl, "max-age=")[1])
		rq.NoError(err, "convert max-age to integer")

		a.LessOrEqual(minMaxAge, age, "minimum max-age >= 3 months")
		a.NotEmpty(rec.Header().Get("Last-Modified"), "should have Last-Modified header")
	})

	t.Run("mdhtml uri returns 404 when non-files requested", func(*testing.T) {
		r := router.NewRouter()
		HandleData(r, tmpDir, middleware.AuthGuard(u))
		_ = os.WriteFile(tmpDir+"/blog/public/0987654321.mdhtml", []byte("i am mdhtml"), 0o600)

		rec := testutils.MockRequest(
			r.Handler,
			"GET",
			"/data/blog/public/0987654321",
			&map[string][]string{
				"Authorization": {"Bearer " + userID},
			},
		)

		a.Equal(rec.Code, http.StatusNotFound, "expect not found")
	})

	t.Run("summary uri protects redeem routes", func(*testing.T) {
		r := router.NewRouter()
		HandleData(r, tmpDir, middleware.AuthGuard(u))

		rec := testutils.MockRequest(
			r.Handler,
			"GET",
			"/data/blog/red33m",
			&map[string][]string{
				"Authorization": {"Bearer " + userID},
			},
		)

		a.Equal(rec.Code, http.StatusUnauthorized, "expect status unauthorized")

		rec = testutils.MockRequest(
			r.Handler,
			"GET",
			"/data/blog/red33m.json",
			&map[string][]string{
				"Authorization": {"Bearer " + userID},
			},
		)

		a.Equal(
			rec.Code,
			http.StatusUnauthorized,
			"expect unauthorized before 'not found' status",
		)
	})

	t.Run("mdhtml uri protects redeem routes", func(*testing.T) {
		r := router.NewRouter()
		_ = os.WriteFile(tmpDir+"/blog/red33m/12345.mdhtml", []byte("test text"), 0o600)
		HandleData(r, tmpDir, middleware.AuthGuard(u))

		rec := testutils.MockRequest(
			r.Handler,
			"GET",
			"/data/blog/red33m/12345.mdhtml",
			&map[string][]string{
				"Authorization": {"Bearer " + userID},
			},
		)

		a.Equal(rec.Code, http.StatusUnauthorized, "expect status unauthorized")

		rec = testutils.MockRequest(
			r.Handler,
			"GET",
			"/data/blog/red33m/12345",
			&map[string][]string{
				"Authorization": {"Bearer " + userID},
			},
		)

		a.Equal(
			rec.Code,
			http.StatusUnauthorized,
			"expect unauthorized before 'not found' status",
		)
	})
}
