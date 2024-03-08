package lib

import (
	"os"
	"sync"
)

func NewFileWriter(filePath string) *FileWriter {
	ch := make(chan ChannelData, 100)
	fw := &FileWriter{
		ch: ch,
	}
	fw.wg.Add(1)
	go fileChannel(fw, filePath, ch)
	return fw
}

func fileChannel(fw *FileWriter, filePath string, ch chan ChannelData) {
	defer fw.wg.Done()
	fileFlag := os.O_CREATE | os.O_RDWR
	activeFile, err := os.OpenFile(filePath, fileFlag, 0o644)
	if err != nil {
		panic(err)
	}

	for chanData := range ch {
		if chanData.IsAppending {
			activeFile.WriteString(chanData.String)
			continue
		}

		err := activeFile.Truncate(0)
		if err != nil {
			panic(err)
		}
		_, err = activeFile.Seek(0, 0)
		if err != nil {
			panic(err)
		}

		_, err = activeFile.WriteString(chanData.String)
		if err != nil {
			panic(err)
		}
	}

	if activeFile != nil {
		err := activeFile.Close()
		if err != nil {
			panic(err)
		}
	}
}

type ChannelData struct {
	String      string
	IsAppending bool
}

type FileWriter struct {
	wg sync.WaitGroup
	ch chan ChannelData
}

func (fa *FileWriter) WriteString(s string, isAppending bool) {
	fa.ch <- ChannelData{
		IsAppending: isAppending,
		String:      s,
	}
}

func (fa *FileWriter) Close() {
	close(fa.ch)
}