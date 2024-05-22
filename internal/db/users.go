package db

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Everything-Explained/go-server/internal"
	"github.com/Everything-Explained/go-server/internal/writers"
)

var ErrUsersClosed = errors.New("user database has been closed")

/*
NewUsers initializes a new user database at the specified
directory.
*/
func NewUsers(dir string) (*Users, error) {
	filePath := dir + "/users.txt"
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return nil, fmt.Errorf("file error::%s", err)
	}

	parsedUsers, err := parseUsers(filePath)
	if err != nil {
		return nil, fmt.Errorf("user parse error::%s", err)
	}

	u := &Users{
		users:      parsedUsers,
		fileWriter: writers.NewFileWriter(f),
		resumeCh:   make(chan bool, 1),
	}

	go u.saveRoutine(500)
	return u, nil
}

type Users struct {
	sync.Mutex
	users          map[string]bool
	fileWriter     *writers.FileWriter
	resumeCh       chan bool
	isPaused       bool
	lastSavedMilli int64
	isClosed       bool
}

/*
Close closes the channels and file writer associated with
the users struct.
*/
func (u *Users) Close() {
	u.Lock()
	u.resumeCh <- false
	close(u.resumeCh)
	u.fileWriter.Close()
	u.fileWriter.WaitGroup.Wait()
	u.isClosed = true
	u.Unlock()
}

func (u *Users) Add(isRed33m bool) (string, error) {
	u.Lock()
	if u.isClosed {
		u.Unlock()
		return "", ErrUsersClosed
	}
	u.Unlock()

	var red33mState bool
	if isRed33m {
		red33mState = true
	}
	newID := internal.GetLongID()
	now := time.Now().UnixMilli()

	u.Lock()
	defer u.Unlock()
	u.users[newID] = red33mState
	u.lastSavedMilli = now
	if u.isPaused {
		u.isPaused = false
		u.resumeCh <- true
	}
	return newID, nil
}

func (u *Users) Update(userid string, isRed33m bool) error {
	u.Lock()
	if u.isClosed {
		u.Unlock()
		return ErrUsersClosed
	}
	u.Unlock()

	var red33mState bool
	if isRed33m {
		red33mState = true
	}

	u.Lock()
	defer u.Unlock()

	if _, exists := u.users[userid]; exists {
		u.users[userid] = red33mState
		if u.isPaused {
			u.isPaused = false
			u.resumeCh <- true
		}
	}
	return nil
}

func (u *Users) Clean() (int, error) {
	u.Lock()
	if u.isClosed {
		u.Unlock()
		return 0, ErrUsersClosed
	}
	u.Unlock()

	delCount := 0
	var sb strings.Builder

	u.Lock()
	defer u.Unlock()

	for k, v := range u.users {
		if !v {
			delCount++
			delete(u.users, k)
			continue
		}
		_, _ = sb.WriteString(fmt.Sprintf("%s: %d\n", k, 1))
	}
	u.fileWriter.WriteString(sb.String(), false)
	return delCount, nil
}

func (u *Users) GetState(userid string) (bool, error) {
	u.Lock()
	if u.isClosed {
		u.Unlock()
		return false, ErrUsersClosed
	}
	userState, exists := u.users[userid]
	u.Unlock()

	if !exists {
		return false, errors.New("user not found")
	}

	return userState, nil
}

func (u *Users) GetLength() int {
	return len(u.users)
}

func (u *Users) GetRandomUserId() (string, error) {
	randIdx := rand.Intn(len(u.users)) // #nosec G404 -- not applicable
	count := 0
	for k := range u.users {
		if count == randIdx {
			return k, nil
		}
		count++
	}
	return "", fmt.Errorf(
		"invalid range::user length: %d, random index: %d",
		len(u.users),
		randIdx,
	)
}

/*
saveRoutine saves users on a cycle, to allow concurrent writes to the user
map without deadlocks.
*/
func (u *Users) saveRoutine(saveDelay uint16) {
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
			isResuming := <-u.resumeCh
			if !isResuming {
				break
			}
		}

		u.Lock()
		for k, v := range u.users {
			var userState byte
			if v {
				userState = 1
			}
			_, _ = sb.WriteString(fmt.Sprintf("%s: %d\n", k, userState))
		}
		u.fileWriter.WriteString(sb.String(), false)
		lastWriteMilli = u.lastSavedMilli
		u.isPaused = false
		u.Unlock()
	}
}

func parseUsers(path string) (map[string]bool, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return make(map[string]bool), err
	}

	users := make(map[string]bool)
	if len(f) == 0 {
		return users, nil
	}

	// === File Format ===
	// string: byte\n
	//
	userArray := strings.Split(strings.TrimSpace(string(f)), "\n")
	for i := range userArray {
		userData := strings.Split(userArray[i], ": ")
		userAccess, err := strconv.Atoi(userData[1])
		if err != nil {
			return make(map[string]bool), err
		}
		users[userData[0]] = byte(userAccess) == 1
	}

	return users, nil
}
