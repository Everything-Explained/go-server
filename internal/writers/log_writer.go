package writers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Everything-Explained/go-server/internal"
)

const (
	logFolderPath = "./logs"
	separator     = "<|>"
)

type LogLevel byte

const (
	DEBUG LogLevel = iota
	INFO
	ERROR
)

var (
	logs map[string]*FileWriter = map[string]*FileWriter{}
	Log  logger                 = logger{}
)

/*
NewLogWriter initializes a new log file with the specified name.
*/
func NewLogWriter(name string) error {
	if _, exists := logs[name]; exists {
		return fmt.Errorf("log name (%s) already exists", name)
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

func (lgr logger) Debug(name string, messages ...any) {
	log(name, DEBUG, messages...)
}

func (lgr logger) Info(name string, messages ...any) {
	log(name, INFO, messages...)
}

func (lgr logger) Error(name string, messages ...any) {
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
		fmt.Sprintf("%dms%d%s%s%s\n", now, level, separator, id, buildLog(messages...)),
		true,
	)
}

func buildLog(messages ...any) string {
	sb := strings.Builder{}
	for _, msg := range messages {
		sb.WriteString(fmt.Sprintf("%s%v", separator, msg))
	}
	return sb.String()
}
