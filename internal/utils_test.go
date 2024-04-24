package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkingDir(t *testing.T) {
	t.Parallel()
	t.Run("gets working directory", func(*testing.T) {
		want, err := os.Getwd()
		assert.NoError(t, err, "get working directory")
		assert.Equal(t, want, Getwd())
	})
}

func TestID(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	t.Run("long id", func(*testing.T) {
		a.Greater(
			len(GetLongID()),
			20,
			"long id should be at least canonical length (21)",
		)
	})

	t.Run("short id", func(*testing.T) {
		a.Less(
			len(GetShortID()),
			21,
			"short id should be less than canonical length (21)",
		)
	})

	t.Run("LengthDiff", func(*testing.T) {
		a.GreaterOrEqual(
			len(GetLongID())-len(GetShortID()),
			5,
			"min distance between short & long id",
		)
	})
}
