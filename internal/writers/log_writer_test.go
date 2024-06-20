package writers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
TestLogWriter only tests the parts of the code that are not
already tested by our integration tests: /middleware/req_logger_test.go
*/
func TestLogWriter(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	r := require.New(t)

	tmpDir := t.TempDir()

	t.Run("should create log file", func(*testing.T) {
		err := CreateLog("log1", tmpDir)
		r.NoError(err, "create log file")
		defer Log.Close("log1")

		f, err := os.OpenFile(tmpDir+"/log1.txt", os.O_CREATE|os.O_APPEND, 0o644)
		r.NoError(err, "open log file")
		defer f.Close()
	})

	t.Run("should append to log file", func(*testing.T) {
		err := CreateLog("log2", tmpDir)
		r.NoError(err, "create log file")
		defer Log.Close("log2")

		Log.Info("log2", "hello world")
		Log.Info("log2", "hello again")

		fileData := openFile(t, tmpDir+"/log2.txt")
		a.Contains(fileData, "hello world")
		a.Contains(fileData, "hello again")
	})

	t.Run("should panic if log name does not exist", func(*testing.T) {
		a.PanicsWithValue(
			"missing log named: 'log'",
			func() {
				Log.Info("log", "hello world")
			},
		)
	})
}
