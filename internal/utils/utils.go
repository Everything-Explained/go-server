package utils

import (
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	workingDir string
	once       sync.Once
)

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
