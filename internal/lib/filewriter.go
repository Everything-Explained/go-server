package lib

import (
	"os"
	"sync"
)

func NewFileWriter(filePath string) *FileWriter {
	var ch = make(chan ChannelData, 100)
	fw := &FileWriter{
		ch: ch,
	}
	fw.wg.Add(1)

	go func() {
		defer fw.wg.Done()
		for chanData := range ch {
			fileFlag := os.O_CREATE | os.O_WRONLY
			if chanData.IsAppending {
				fileFlag = fileFlag | os.O_APPEND
			}

			f, err := os.OpenFile(filePath, fileFlag, 0644)
			if err != nil {
				panic(err)
			}
			f.WriteString(chanData.String)
			f.Close()
		}

	}()

	return fw
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
