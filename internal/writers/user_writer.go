package writers

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Everything-Explained/go-server/internal"
)

var userFile string = internal.Getwd() + "/users.txt"

var user *UserWriter

func User() *UserWriter {
	if user == nil {
		f, err := os.OpenFile(userFile, os.O_WRONLY|os.O_CREATE, 0o644)
		if err != nil {
			panic(err)
		}

		user = &UserWriter{
			fileWriter: NewFileWriter(f),
			users:      parseUsers(userFile),
			resumeCh:   make(chan bool, 1),
		}

		go user.saveRoutine(500)
	}

	return user
}

type UserWriter struct {
	sync.Mutex
	fileWriter     *FileWriter
	resumeCh       chan bool
	isPaused       bool
	users          map[string]byte
	lastSavedMilli int64
}

func (u *UserWriter) Add(isRed33m bool) string {
	var userState byte
	if isRed33m {
		userState = 1
	}
	newID := internal.GetLongID()
	now := time.Now().UnixMilli()
	u.Lock()
	u.users[newID] = userState
	u.lastSavedMilli = now
	if u.isPaused {
		u.isPaused = false
		u.resumeCh <- true
	}
	u.Unlock()
	return newID
}

func (u *UserWriter) GetState(userid string) (byte, error) {
	u.Lock()
	userState, exists := u.users[userid]
	u.Unlock()

	if !exists {
		return 0, errors.New("user not found")
	}

	return userState, nil
}

func (u *UserWriter) Update(userid string, isRed33m bool) {
	var userState byte
	if isRed33m {
		userState = 1
	}
	u.Lock()
	if _, exists := u.users[userid]; exists {
		u.users[userid] = userState
		if u.isPaused {
			u.isPaused = false
			u.resumeCh <- true
		}
	}
	u.Unlock()
}

func (u *UserWriter) Close() {
	close(u.fileWriter.ch)
	u.fileWriter.wg.Wait()
}

/*
saveRoutine saves users on a cycle, to allow concurrent writes to the user
map without deadlocks.
*/
func (u *UserWriter) saveRoutine(saveDelay uint16) {
	if saveDelay < 30 {
		panic("save delay must be at least 30 milliseconds")
	}

	var lastWriteMilli int64

	for {
		sb := strings.Builder{}
		time.Sleep(time.Duration(saveDelay) * time.Millisecond)

		u.Lock()
		isPausing := u.lastSavedMilli == lastWriteMilli
		u.isPaused = isPausing
		u.Unlock()

		if isPausing {
			<-u.resumeCh
		}

		u.Lock()
		for k, v := range u.users {
			_, _ = sb.WriteString(fmt.Sprintf("%s: %d\n", k, v))
		}
		u.fileWriter.WriteString(sb.String(), false)
		lastWriteMilli = u.lastSavedMilli
		u.isPaused = false
		u.Unlock()
	}
}

func parseUsers(filePath string) map[string]byte {
	f, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	users := make(map[string]byte)
	if len(f) == 0 {
		return users
	}

	// === File Format ===
	// string: byte\n
	//
	userArray := strings.Split(strings.TrimSpace(string(f)), "\n")
	for i := range userArray {
		userData := strings.Split(userArray[i], ": ")
		userAccess, err := strconv.Atoi(userData[1])
		if err != nil {
			panic(err)
		}
		users[userData[0]] = byte(userAccess)
	}

	return users
}
