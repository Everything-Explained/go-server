package lib

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Everything-Explained/go-server/internal/utils"
	"github.com/oklog/ulid/v2"
)

var userFile string = utils.GetWorkingDir() + "\\users.txt"

var UserWriter = &userWriter{
	fw:   NewFileWriter(userFile),
	data: parseUsers(userFile),
}

type userWriter struct {
	sync.Mutex
	fw       *FileWriter
	data     map[string]uint8
	isSaving bool
}

func (u *userWriter) AddUser(isRed33m bool) {
	var userState uint8 = 0
	if isRed33m {
		userState = 1
	}
	newID := ulid.Make().String()
	u.Lock()
	u.data[newID] = userState
	u.fw.WriteString(fmt.Sprintf("%s: %b\n", newID, userState), true)
	u.Unlock()
}

func (u *userWriter) GetUserState(id string) (uint8, error) {
	u.Lock()
	userState, exists := u.data[id]
	u.Unlock()

	if !exists {
		return 0, errors.New("user not found")
	}

	return userState, nil
}

func (u *userWriter) UpdateUser(id string, isRed33m bool) {
	var userState uint8
	if isRed33m {
		userState = 1
	}
	u.Lock()
	if _, exists := u.data[id]; exists {
		u.data[id] = userState
		if !u.isSaving {
			go saveUsers(u)
		}
	}
	u.Unlock()
}

func (u *userWriter) Close() {
	close(u.fw.ch)
	u.fw.wg.Wait()
}

// saveUsers delays the overwriting of the user file so that
// other save operations can continue to update the user
// map without excessive expensive writes.
func saveUsers(u *userWriter) {
	u.Lock()
	u.isSaving = true
	u.Unlock()

	time.Sleep(500 * time.Millisecond)
	var sb strings.Builder

	u.Lock()
	sb.Grow(len(u.data))
	for k, v := range u.data {
		sb.WriteString(fmt.Sprintf("%s: %d\n", k, v))
	}
	u.fw.WriteString(sb.String(), false)
	u.isSaving = false
	u.Unlock()
}

func parseUsers(filePath string) map[string]uint8 {
	f, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	users := make(map[string]uint8)
	if len(f) == 0 {
		return users
	}

	// FORMAT: string: uint8\n
	userArray := strings.Split(strings.TrimSpace(string(f)), "\n")
	for i := 0; i < len(userArray); i++ {
		userData := strings.Split(userArray[i], ": ")
		userAccess, err := strconv.Atoi(userData[1])
		if err != nil {
			panic(err)
		}
		users[userData[0]] = uint8(userAccess)
	}

	return users
}
