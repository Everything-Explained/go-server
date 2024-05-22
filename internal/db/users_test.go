package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsers(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	t.Run("Inits user txt file", func(*testing.T) {
		dir := t.TempDir()

		u, err := NewUsers(dir)
		a.NoError(err, "should create new users dir")

		b, err := os.ReadFile(dir + "/users.txt")
		a.NoError(err, "should find users text file")
		a.Equal(b, []byte{})

		u.Close()
	})

	t.Run("adds user to users txt file", func(*testing.T) {
		dir := t.TempDir()
		u, err := NewUsers(dir)
		a.NoError(err, "should create new users dir")

		usrStr, err := u.Add(false)
		a.NoError(err, "should add user")
		a.Len(usrStr, 24, "should be a 24 char hash")
		u.Close()
	})

	t.Run("sets & gets user state", func(*testing.T) {
		dir := t.TempDir()
		u, err := NewUsers(dir)
		a.NoError(err, "should create new users dir")
		defer u.Close()

		usrStr, err := u.Add(false)
		a.NoError(err, "should add user")
		state, err := u.GetState(usrStr)
		a.NoError(err, "should return state")
		a.Equal(state, false, "user should NOT be red33med")

		err = u.Update(usrStr, true)
		a.NoError(err, "should update user state")
		state, err = u.GetState(usrStr)
		a.NoError(err, "should return state")
		a.Equal(state, true, "user should be red33med")
	})

	t.Run("gets length of users", func(*testing.T) {
		dir := t.TempDir()
		u, err := NewUsers(dir)
		a.NoError(err, "should create new users dir")
		defer u.Close()

		for range 10 {
			_, err := u.Add(false)
			a.NoError(err, "should add user")
		}

		a.Equal(u.GetLength(), 10, "there should be 10 users")
	})
}
