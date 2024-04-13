package internal

import (
	"os"
	"testing"

	"github.com/Everything-Explained/go-server/testutils"
)

func TestWorkingDir(t *testing.T) {
	t.Run("gets working directory", func(t *testing.T) {
		got := Getwd()
		want, _ := os.Getwd()
		if got != want {
			t.Error(testutils.PrintErrorS(got, want))
		}
	})

	t.Run("gets active working dir", func(t *testing.T) {
		err := os.Chdir("../")
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		got := Getwd()
		want, _ := os.Getwd()
		if got != want {
			t.Error(testutils.PrintErrorS(want, got))
		}
	})
}

func TestID(t *testing.T) {
	t.Parallel()

	t.Run("Long", func(t *testing.T) {
		got := len(GetLongID())

		if got < 21 {
			want := "long IDs should be at least, canonical length (21)"
			t.Error(testutils.PrintErrorS(want, got))
		}
	})

	t.Run("Short", func(t *testing.T) {
		got := len(GetShortID())

		if got >= 21 {
			want := "short IDs should be less than canonical length (21)"
			t.Error(testutils.PrintErrorS(want, got))
		}
	})

	t.Run("LengthDiff", func(t *testing.T) {
		shortLen := len(GetShortID())
		longLen := len(GetLongID())

		if longLen-shortLen < 5 {
			ex := "min distance between short & long IDs is 5"
			testutils.PrintErrorD(ex, shortLen, longLen)
		}
	})
}
