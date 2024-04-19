package routes

import (
	"net/http"
	"os"
	"testing"

	"github.com/Everything-Explained/go-server/internal"
	"github.com/Everything-Explained/go-server/internal/db"
	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/testutils"
)

type Mock403 struct {
	Headers *map[string][]string
}

func TestSetupRoute(t *testing.T) {
	reset := testutils.SetTempDir(t)
	defer reset()

	r := router.NewRouter()
	err := os.WriteFile("mock.txt", []byte("test text"), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	HandleSetup(r, internal.Getwd()+"/mock.txt")
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

			if resp.Code != http.StatusForbidden {
				t.Error(testutils.PrintErrorS(http.StatusForbidden, resp.Code))
			}

			if resp.Body.String() != expBody {
				t.Error(testutils.PrintErrorS(expBody, resp.Body.String()))
			}

		}
	})

	t.Run("detects unauthorized request", func(t *testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/setup", &map[string][]string{
			"Authorization": {"Bearer gibberish"},
		})

		if resp.Code != http.StatusUnauthorized {
			t.Log(resp.Body.String())
			t.Error(testutils.PrintErrorS(http.StatusUnauthorized, resp.Code))
		}

		expBody := "Authorization Expired or Missing\n"
		if resp.Body.String() != expBody {
			t.Error(testutils.PrintErrorS(expBody, resp.Body.String()))
		}
	})

	t.Run("adds users", func(t *testing.T) {
		resp := testutils.MockRequest(r.Handler, "GET", "/setup", &map[string][]string{
			"Authorization": {"Bearer setup"},
		})

		if resp.Code != http.StatusOK {
			t.Error(testutils.PrintErrorS(http.StatusOK, resp.Code))
		}

		id := resp.Header().Get("X-Evex-Id")
		if id == "" {
			t.Error(
				testutils.PrintErrorD(
					"Should have X-Evex-Id header",
					"header exists",
					"header does not exist",
				),
			)
		}

		users := db.GetUsers()
		defer users.Close()

		isRed33med, err := users.GetState(id)
		if err != nil {
			t.Error(testutils.PrintErrorD("Return the user state", "user to exist", err))
		}

		if isRed33med {
			t.Error(
				testutils.PrintErrorD(
					"Users should not have red33m access by default",
					false,
					isRed33med,
				),
			)
		}

		if resp.Body.String() != "test text" {
			t.Error(testutils.PrintErrorS("test text", resp.Body.String()))
		}
	})

	t.Run("detects authenticated user", func(t *testing.T) {
		id, _ := db.GetUsers().GetRandomUserId()
		resp := testutils.MockRequest(r.Handler, "GET", "/setup", &map[string][]string{
			"Authorization": {"Bearer " + id},
		})

		if resp.Code != http.StatusOK {
			t.Error(testutils.PrintErrorS(http.StatusOK, resp.Code))
		}

		if resp.Body.String() != "test text" {
			t.Error(
				testutils.PrintErrorD(
					"Should have expected body value",
					"test text",
					resp.Body.String(),
				),
			)
		}

		idHeader := resp.Header().Get("X-Evex-Id")
		if idHeader != "" {
			t.Error(
				testutils.PrintErrorD("Should not include ID header", "empty string", idHeader),
			)
		}

		red33mVal := resp.Header().Get("X-Evex-Red33m")
		if red33mVal != "no" {
			t.Error(
				testutils.PrintErrorD(
					"Should have expected red33m header value",
					"no",
					red33mVal,
				),
			)
		}
	})

	t.Run("detects redeem user", func(t *testing.T) {
		u := db.GetUsers()
		id, _ := u.GetRandomUserId()
		u.Update(id, true)
		resp := testutils.MockRequest(r.Handler, "GET", "/setup", &map[string][]string{
			"Authorization": {"Bearer " + id},
		})

		if resp.Code != http.StatusOK {
			t.Error(testutils.PrintErrorS(http.StatusOK, resp.Code))
		}

		headerVal := resp.Header().Get("X-Evex-Red33m")
		if headerVal != "yes" {
			t.Error(
				testutils.PrintErrorD(
					"Should have red33m header value",
					"yes",
					headerVal,
				),
			)
		}
	})
}
