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

var users *Users

/*
CreateUsers initializes the users database at the specified
directory. This function can only be effectively called
once, since it initializes a singleton.

📝 Subsequent calls to CreateUsers() are effectively ignored.
*/
func CreateUsers(dir string) error {
	if users == nil {
		filePath := dir + "/users.txt"
		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0o644)
		if err != nil {
			return fmt.Errorf("file error::%s", err)
		}

		go saveRoutine(500)

		parsedUsers, err := parseUsers(filePath)
		if err != nil {
			return fmt.Errorf("user parse error::%s", err)
		}

		users = &Users{
			users:      parsedUsers,
			fileWriter: writers.NewFileWriter(f),
			resumeCh:   make(chan bool, 1),
		}
	}
	return nil
}

func GetUsers() *Users {
	checkIfCreated()
	return users
}

type Users struct {
	sync.Mutex
	users          map[string]bool
	fileWriter     *writers.FileWriter
	resumeCh       chan bool
	isPaused       bool
	lastSavedMilli int64
}

/*
Close closes the channels and file writer associated with
the users struct.

🟠 Continuing to use the users struct after it has been
closed will cause unexpected behavior.
*/
func (u *Users) Close() {
	checkIfCreated()
	close(u.fileWriter.Channel)
	close(u.resumeCh)
	u.fileWriter.WaitGroup.Wait()
}

func (u *Users) Add(isRed33m bool) string {
	checkIfCreated()
	var red33mState bool
	if isRed33m {
		red33mState = true
	}
	newID := internal.GetLongID()
	now := time.Now().UnixMilli()
	u.Lock()
	u.users[newID] = red33mState
	u.lastSavedMilli = now
	if u.isPaused {
		u.isPaused = false
		u.resumeCh <- true
	}
	u.Unlock()
	return newID
}

func (u *Users) Update(userid string, isRed33m bool) {
	checkIfCreated()
	var red33mState bool
	if isRed33m {
		red33mState = true
	}
	u.Lock()
	if _, exists := u.users[userid]; exists {
		u.users[userid] = red33mState
		if u.isPaused {
			u.isPaused = false
			u.resumeCh <- true
		}
	}
	u.Unlock()
}

func (u *Users) Clean() int {
	checkIfCreated()
	delCount := 0
	var sb strings.Builder
	u.Lock()
	for k, v := range u.users {
		if !v {
			delCount++
			delete(u.users, k)
			continue
		}
		_, _ = sb.WriteString(fmt.Sprintf("%s: %d\n", k, 1))
	}
	u.fileWriter.WriteString(sb.String(), false)
	u.Unlock()
	return delCount
}

func (u *Users) GetState(userid string) (bool, error) {
	checkIfCreated()
	u.Lock()
	userState, exists := u.users[userid]
	u.Unlock()

	if !exists {
		return false, errors.New("user not found")
	}

	return userState, nil
}

func (u *Users) GetLength() int {
	checkIfCreated()
	return len(u.users)
}

func (u *Users) GetRandomUserId() (string, error) {
	checkIfCreated()
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

func checkIfCreated() {
	if users == nil {
		panic("users not initialized; did you forget to create it?")
	}
}

/*
saveRoutine saves users on a cycle, to allow concurrent writes to the user
map without deadlocks.
*/
func saveRoutine(saveDelay uint16) {
	if saveDelay < 30 {
		panic("save delay must be at least 30 milliseconds")
	}

	var lastWriteMilli int64

	for {
		sb := strings.Builder{}
		time.Sleep(time.Duration(saveDelay) * time.Millisecond)

		u := GetUsers()
		u.Lock()
		isPausing := u.lastSavedMilli == lastWriteMilli
		u.isPaused = isPausing
		u.Unlock()

		if isPausing {
			<-u.resumeCh
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
