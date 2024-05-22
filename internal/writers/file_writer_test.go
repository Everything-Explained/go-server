package writers

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileWriter(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	mockFile := createFileMocker(t)

	t.Run("appends to file", func(*testing.T) {
		fw, filePath := mockFile(os.O_APPEND)
		fw.WriteString("abcdefghijklmn", true)
		fw.WriteString("opqrstuv", true)
		fw.WriteString("wxyz", true)
		<-fw.OutChan
		<-fw.OutChan
		<-fw.OutChan
		fw.Close()

		data := openFile(t, filePath)
		a.Equal(
			"test text\nabcdefghijklmnopqrstuvwxyz",
			data,
			"should append text to file",
		)
	})

	t.Run("overwrites file contents", func(*testing.T) {
		fw, filePath := mockFile(os.O_WRONLY)
		defer fw.Close()
		fw.WriteString("appended", true)
		<-fw.OutChan

		data := openFile(t, filePath)
		a.Equal("test text\nappended", data, "should append data")

		fw.WriteString("overwriting file", false)
		<-fw.OutChan

		data = openFile(t, filePath)
		a.Equal("overwriting file", data, "should overwrite file contents")
	})

	t.Run("can gracefully close", func(*testing.T) {
		fw, _ := mockFile(os.O_APPEND)
		a.NotNil(fw.inChan, "input channel should be available")
		a.NotNil(fw.OutChan, "output channel should be available")
		fw.Close()
		_, inChanState := <-fw.inChan
		_, outChanState := <-fw.OutChan
		a.False(inChanState, "input channel should be closed")
		a.False(outChanState, "output channel should be closed")
	})
}

func createFileMocker(t *testing.T) func(int) (*FileWriter, string) {
	count := 0
	return func(flags int) (*FileWriter, string) {
		count++
		d := t.TempDir()
		filePath := fmt.Sprintf("%s/mock%d.txt", d, count)
		f, err := os.OpenFile(filePath, os.O_CREATE|flags, 0o644)
		require.NoError(t, err, "create mock file")

		_, err = f.WriteString("test text\n")
		require.NoError(t, err, "write mock data")

		return NewFileWriter(f, true), filePath
	}
}

func openFile(t *testing.T, filePath string) string {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDONLY, 0o644)
	require.NoError(t, err, "open mock file")
	defer f.Close()

	data, err := io.ReadAll(f)
	require.NoError(t, err, "read mock file data")

	err = f.Close()
	require.NoError(t, err, "close mock file")
	return string(data)
}
