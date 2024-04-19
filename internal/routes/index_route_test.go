package routes

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/testutils"
)

func TestIndexRoute(t *testing.T) {
	reset := testutils.SetTempDir(t)
	defer reset()

	r := router.NewRouter()
	err := os.WriteFile("mock.html", []byte("index text"), 0o600)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	HandleIndex(r, "./mock.html")

	resp := testutils.MockRequest(r.Handler, "GET", "/", nil)

	t.Run("has 200 status", func(t *testing.T) {
		if resp.Code != http.StatusOK {
			t.Error(testutils.PrintErrorS(http.StatusOK, resp.Code))
		}
	})

	t.Run("has html content type", func(t *testing.T) {
		got := resp.Result().Header.Get("Content-Type")
		want := "text/html; charset=utf-8"

		if got != want {
			t.Error(testutils.PrintErrorS(want, got))
		}
	})

	t.Run("returns index when default url", func(t *testing.T) {
		data, _ := io.ReadAll(resp.Body)
		got := strings.TrimSpace(string(data))
		want := "index text"

		if got != want {
			t.Error(testutils.PrintErrorS(want, got))
		}
	})

	t.Run("returns index when index.html", func(t *testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/index.html", nil)

		wantStatus := http.StatusOK
		gotStatus := resp.Code

		if wantStatus != gotStatus {
			t.Error(testutils.PrintErrorS(wantStatus, gotStatus))
		}

		d, _ := io.ReadAll(resp.Body)
		gotBody := strings.TrimSpace(string(d))
		wantBody := "index text"

		if wantBody != gotBody {
			t.Error(testutils.PrintErrorS(wantBody, gotBody))
		}
	})

	t.Run("returns index when unknown url", func(t *testing.T) {
		wantStatus := http.StatusOK
		gotStatus := resp.Code

		if wantStatus != gotStatus {
			t.Error(testutils.PrintErrorS(wantStatus, gotStatus))
		}
	})
}
