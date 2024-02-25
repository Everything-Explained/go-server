package utils

import (
	"os"
	"time"
)

var WorkingDir string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	WorkingDir = wd
}

func GetISODateNow() string {
	const ISODate8601Format = "2006-01-02T15:04:05.000Z07:00:00"
	return time.Now().UTC().Format(ISODate8601Format)
}
