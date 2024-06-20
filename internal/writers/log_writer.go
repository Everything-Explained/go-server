package writers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Everything-Explained/go-server/internal"
)

const (
	separator   string = "<|>"
	newLineChar string = "\u200B"
)

type LogLevel byte

const (
	DEBUG LogLevel = iota
	INFO
	ERROR
)

var (
	logs = make(map[string]*FileWriter)
	Log  = Logger{}
)

/*
CreateLog initializes a new log file with the specified name and
directory.
*/
func CreateLog(name string, logDir string) error {
	if _, exists := logs[name]; exists {
		return nil
	}

	err := os.MkdirAll(logDir, 0o755)
	if err != nil {
		return err
	}

	logFilePath := logDir + "/" + name + ".txt"
	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	logs[name] = NewFileWriter(f, false)
	return nil
}

type Logger struct{}

func (Logger) Debug(name string, messages ...any) {
	log(name, DEBUG, messages...)
}

func (Logger) Info(name string, messages ...any) {
	log(name, INFO, messages...)
}

func (Logger) Error(name string, messages ...any) {
	log(name, ERROR, messages...)
}

func (Logger) Close(name string) {
	f, ok := logs[name]
	if !ok {
		panic(fmt.Errorf("log not found: %s", name))
	}
	f.Close()
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
