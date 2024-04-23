package routes

import (
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/testutils"
)

func TestAssetRoute(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	err := os.Mkdir(dir+"/mocks", 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(dir+"/mocks/mock.txt", []byte("test text"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	HandleAssets(r, dir+"/mocks")

	rec := testutils.MockRequest(r.Handler, "GET", "/assets/mock.txt", nil)

	t.Run("status 200", func(t *testing.T) {
		if rec.Code != http.StatusOK {
			t.Error(testutils.PrintErrorS(http.StatusOK, rec.Code))
		}
	})

	t.Run("returns file", func(t *testing.T) {
		d, _ := io.ReadAll(rec.Body)
		gotBody := string(d)
		wantBody := "test text"

		if gotBody != wantBody {
			t.Error(testutils.PrintErrorS(wantBody, gotBody))
		}
	})
}
