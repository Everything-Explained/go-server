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
	reset := testutils.SetTempDir(t)
	defer reset()

	err := os.Mkdir("./mocks", 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile("./mocks/mock.txt", []byte("test text"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	r := router.NewRouter()
	HandleAssets(r, "./mocks")

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
