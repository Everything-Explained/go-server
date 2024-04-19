package writers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Everything-Explained/go-server/internal"
)

const (
	logFolderPath string = "./logs"
	separator     string = "<|>"
	newLineChar   string = "\u200B"
)

type LogLevel byte

const (
	DEBUG LogLevel = iota
	INFO
	ERROR
)

var (
	logs = make(map[string]*FileWriter)
	Log  = logger{}
)

/*
NewLogWriter initializes a new log file with the specified name.
*/
func NewLogWriter(name string) error {
	if _, exists := logs[name]; exists {
		return nil
	}

	err := os.MkdirAll(logFolderPath, 0o755)
	if err != nil {
		return err
	}

	logFilePath := logFolderPath + "/" + name + ".txt"
	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	logs[name] = NewFileWriter(f)
	return nil
}

type logger struct{}

func (logger) Debug(name string, messages ...any) {
	log(name, DEBUG, messages...)
}

func (logger) Info(name string, messages ...any) {
	log(name, INFO, messages...)
}

func (logger) Error(name string, messages ...any) {
	log(name, ERROR, messages...)
}

func log(name string, level LogLevel, messages ...any) {
	l, exists := logs[name]
	if !exists {
		panic(fmt.Errorf("missing log named: '%s'", name))
	}

	id := internal.GetShortID()
	now := time.Now().UnixMilli()

	l.WriteString(
		fmt.Sprintf("%dms%s%s\n", now, buildLog(level, id), buildLog(messages...)),
		true,
	)
}

func buildLog(messages ...any) string {
	sb := strings.Builder{}
	for _, msg := range messages {
		_, _ = sb.WriteString(fmt.Sprintf("%s%v", separator, msg))
	}
	s := sb.String()
	if strings.Contains(s, "\r\n") {
		return strings.ReplaceAll(s, "\r\n", newLineChar)
	}

	if strings.Contains(s, "\n") {
		return strings.ReplaceAll(s, "\n", newLineChar)
	}
	return s
}
