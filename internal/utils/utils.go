package utils

import (
	"net/http"
	"os"
	"time"
)

var workingDir string

func GetWorkingDir() string {
	if workingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		workingDir = wd
	}
	return workingDir
}

func GetISODateNow() string {
	const ISODate8601Format = "2006-01-02T15:04:05.000Z07:00:00"
	return time.Now().UTC().Format(ISODate8601Format)
}

func GetGMTFrom(t time.Time) string {
	return t.UTC().Format(http.TimeFormat)
}
