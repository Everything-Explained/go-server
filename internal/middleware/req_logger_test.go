package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Everything-Explained/go-server/internal/router"
	"github.com/Everything-Explained/go-server/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRequestLogger(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	tmpDir := t.TempDir()

	t.Run("should log to requests file", func(t *testing.T) {
		r := router.NewRouter()
		closeLog, reqLogger := LogRequests(tmpDir+"/logs", "r1")
		defer closeLog()

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		}, reqLogger(0))

		testutils.MockRequest(
			r.Handler,
			"GET",
			"/",
			nil,
			nil,
		)

		file, err := os.ReadFile(tmpDir + "/logs/r1.txt")
		if a.NoError(err) {
			t.Log(string(file))
			a.Greater(len(file), 0, "requests file should have some content")
		}
	})

	t.Run("should contain request info", func(*testing.T) {
		r := router.NewRouter()
		closeLog, reqLogger := LogRequests(tmpDir+"/logs", "r2")
		defer closeLog()

		r.Get("/test/path", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		}, reqLogger(0))

		currentMs := time.Now().UnixMilli()
		testutils.MockRequest(
			r.Handler,
			"GET",
			"/test/path?test=1",
			strings.NewReader("test body"),
			&map[string][]string{
				"User-Agent":   {"custom-agent"},
				"Cf-Ipcountry": {"US"},
			},
		)

		file, err := os.ReadFile(tmpDir + "/logs/r2.txt")
		if a.NoError(err) {
			fileStr := string(file)
			// Has Separator
			a.Contains(fileStr, "<|>")

			infoParts := strings.Split(fileStr, "<|>")

			// Has timestamp
			a.Equal("ms", infoParts[0][len(infoParts[0])-2:], "timestamp should be ms")
			a.Equal(
				fmt.Sprintf("%d", currentMs),
				infoParts[0][:len(infoParts[0])-2],
				"timestamp should match",
			)

			// Has log level
			a.Equal("1", infoParts[1], "log level should be 1")

			// Has ID
			a.Len(infoParts[2], 8, "id should be 8 chars long")

			// Has user agent
			a.Equal("custom-agent", infoParts[3], "user agent should match")

			// Has country
			a.Equal("US", infoParts[4], "country should be US")

			// Has method
			a.Equal("GET", infoParts[5], "method should be GET")

			// Has host
			a.Equal("example.com", infoParts[6], "host should be a test url")

			// Has path
			a.Equal("/test/path", infoParts[7], "path should be /test/path")

			// Has query
			a.Equal("test=1", infoParts[8], "query should be test=1")

			// Has body
			a.Equal("test body", infoParts[9], "body should be test body")

			// Has status
			a.Equal("200", infoParts[10], "status should be 200")

			// Has request speed
			a.Equal("0µs\n", infoParts[11], "request speed should be 0µs")
		}
	})
}
