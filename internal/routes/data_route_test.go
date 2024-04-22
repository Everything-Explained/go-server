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
)

func TestDataRoute(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	_ = os.Mkdir(tmpDir+"/blog", fs.FileMode(os.O_WRONLY))
	_ = os.Mkdir(tmpDir+"/blog/public", fs.FileMode(os.O_WRONLY))
	_ = os.Mkdir(tmpDir+"/blog/red33m", fs.FileMode(os.O_WRONLY))

	u, err := db.NewUsers(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer u.Close()
	u.Add(false)
	userID, err := u.GetRandomUserId()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	t.Run("panic on missing auth middleware", func(t *testing.T) {
		r := router.NewRouter()
		defer testutils.TestPanic(t, "missing auth guard", "missing auth guard")
		HandleData(r, tmpDir)
		testutils.MockRequest(r.Handler, "GET", "/data/blog/public", nil)
	})

	t.Run("passes summary uri spec", func(t *testing.T) {
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

		if rec.Code != http.StatusOK {
			t.Error(testutils.PrintErrorD("correct status code", 200, rec.Code))
		}

		if rec.Body.String() != "test text" {
			t.Error(testutils.PrintErrorD("correct body text", "test text", rec.Body.String()))
		}

		contentType := rec.Header().Get("Content-Type")
		wantedType := "application/json; charset=utf-8"
		if contentType != wantedType {
			t.Error(
				testutils.PrintErrorD(
					"correct content type header",
					wantedType,
					contentType,
				),
			)
		}

		cacheControl := rec.Header().Get("Cache-Control")
		if !strings.Contains(cacheControl, "max-age=") {
			t.Error(testutils.PrintErrorD("cache control age", "max-age", cacheControl))
		}

		// 3 months
		minMaxAge := 60 * 60 * 24 * 30 * 3
		age, err := strconv.Atoi(strings.Split(cacheControl, "max-age=")[1])
		if err != nil {
			t.Fatalf("could not convert max age to integer: %s", cacheControl)
		}

		if age < minMaxAge {
			t.Error(testutils.PrintErrorD("max-age > 3 months", minMaxAge, age))
		}

		if rec.Header().Get("Last-Modified") == "" {
			t.Error(testutils.PrintErrorS("Last-Modified header", "no Last-Modified header"))
		}
	})

	t.Run("summary uri returns 404 when files requested", func(t *testing.T) {
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

		if rec.Code != http.StatusNotFound {
			t.Error(testutils.PrintErrorD("correct status code", http.StatusNotFound, rec.Code))
		}
	})

	t.Run("passes mdhtml uri spec", func(t *testing.T) {
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

		if rec.Code != http.StatusOK {
			t.Error(testutils.PrintErrorD("correct status code", 200, rec.Code))
		}

		if rec.Body.String() != "i am mdhtml" {
			t.Error(testutils.PrintErrorD("correct body text", "test text", rec.Body.String()))
		}

		contentType := rec.Header().Get("Content-Type")
		wantedType := "text/html; charset=utf-8"
		if contentType != wantedType {
			t.Error(
				testutils.PrintErrorD(
					"correct content type header",
					wantedType,
					contentType,
				),
			)
		}

		cacheControl := rec.Header().Get("Cache-Control")
		if !strings.Contains(cacheControl, "max-age=") {
			t.Error(testutils.PrintErrorD("cache control age", "max-age", cacheControl))
		}

		// 3 months
		minMaxAge := 60 * 60 * 24 * 30 * 3
		age, err := strconv.Atoi(strings.Split(cacheControl, "max-age=")[1])
		if err != nil {
			t.Fatalf("could not convert max age to integer: %s", cacheControl)
		}

		if age < minMaxAge {
			t.Error(testutils.PrintErrorD("max-age > 3 months", minMaxAge, age))
		}

		if rec.Header().Get("Last-Modified") == "" {
			t.Error(testutils.PrintErrorS("Last-Modified header", "no Last-Modified header"))
		}
	})

	t.Run("mdhtml uri returns 404 when non-files requested", func(t *testing.T) {
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

		if rec.Code != http.StatusNotFound {
			t.Error(
				testutils.PrintErrorD("a Not Found status code", http.StatusNotFound, rec.Code),
			)
		}
	})

	t.Run("summary uri protects redeem routes", func(t *testing.T) {
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

		if rec.Code != http.StatusUnauthorized {
			t.Error(
				testutils.PrintErrorD(
					"return Forbidden status",
					http.StatusUnauthorized,
					rec.Code,
				),
			)
		}

		rec = testutils.MockRequest(
			r.Handler,
			"GET",
			"/data/blog/red33m.json",
			&map[string][]string{
				"Authorization": {"Bearer " + userID},
			},
		)

		if rec.Code != http.StatusUnauthorized {
			t.Error(
				testutils.PrintErrorD(
					"return Unauthorized before Not Found status",
					http.StatusUnauthorized,
					rec.Code,
				),
			)
		}
	})

	t.Run("mdhtml uri protects redeem routes", func(t *testing.T) {
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

		if rec.Code != http.StatusUnauthorized {
			t.Error(
				testutils.PrintErrorD(
					"return Forbidden status",
					http.StatusUnauthorized,
					rec.Code,
				),
			)
		}

		rec = testutils.MockRequest(
			r.Handler,
			"GET",
			"/data/blog/red33m/12345",
			&map[string][]string{
				"Authorization": {"Bearer " + userID},
			},
		)

		if rec.Code != http.StatusUnauthorized {
			t.Error(
				testutils.PrintErrorD(
					"return Unauthorized before Not Found status",
					http.StatusUnauthorized,
					rec.Code,
				),
			)
		}
	})
}
