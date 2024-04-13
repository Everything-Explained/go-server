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

type ContextKey struct{ Name string }

func init() {
	longIDFunc, _ := nanoid.Standard(24)
	GetLongID = longIDFunc

	shortIDFunc, _ := nanoid.Standard(8)
	GetShortID = shortIDFunc
}

/*
Getwd gets the working directory. Wraps the built-in os.Getwd()
so we can just panic in the extremely rare case that it
returns an error.

🔴 Panics for any error returned by the os.Getwd().
*/
func Getwd() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}

func GetISODateNow() string {
	const ISODate8601Format = "2006-01-02T15:04:05.000Z07:00:00"
	return time.Now().UTC().Format(ISODate8601Format)
}

func GetGMTFrom(t time.Time) string {
	return t.UTC().Format(http.TimeFormat)
}
