package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkingDir(t *testing.T) {
	t.Parallel()
	t.Run("gets working directory", func(t *testing.T) {
		want, err := os.Getwd()
		assert.NoError(t, err, "get working directory")
		assert.Equal(t, want, Getwd())
	})
}

func TestID(t *testing.T) {
	t.Parallel()

	t.Run("long id", func(t *testing.T) {
		assert.Greater(
			t,
			len(GetLongID()),
			20,
			"long id should be at least canonical length (21)",
		)
	})

	t.Run("short id", func(t *testing.T) {
		assert.Less(
			t,
			len(GetShortID()),
			21,
			"short id should be less than canonical length (21)",
		)
	})

	t.Run("LengthDiff", func(t *testing.T) {
		assert.GreaterOrEqual(
			t,
			len(GetLongID())-len(GetShortID()),
			5,
			"min distance between short & long id",
		)
	})
}
