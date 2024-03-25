package internal

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/jaevor/go-nanoid"
)

var (
	workingDir string
	once       sync.Once
	GetLongID  func() string
	GetShortID func() string
)

func init() {
	longIDFunc, _ := nanoid.Standard(24)
	GetLongID = longIDFunc

	shortIDFunc, _ := nanoid.Standard(8)
	GetShortID = shortIDFunc
}

func GetWorkingDir() string {
	once.Do(func() {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		workingDir = wd
	})
	return workingDir
}

func GetISODateNow() string {
	const ISODate8601Format = "2006-01-02T15:04:05.000Z07:00:00"
	return time.Now().UTC().Format(ISODate8601Format)
}

func GetGMTFrom(t time.Time) string {
	return t.UTC().Format(http.TimeFormat)
}
