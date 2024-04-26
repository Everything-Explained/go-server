package writers

import (
	"os"
	"sync"
)

func NewFileWriter(file *os.File) *FileWriter {
	ch := make(chan ChannelData, 1000)
	fw := &FileWriter{
		Channel: ch,
		file:    file,
	}
	fw.WaitGroup.Add(1)
	go fileChannel(fw, file)
	return fw
}

type ChannelData struct {
	String      string
	IsAppending bool
	close       bool
}

type FileWriter struct {
	WaitGroup sync.WaitGroup
	Channel   chan ChannelData
	file      *os.File
}

func (fa *FileWriter) WriteString(s string, isAppending bool) {
	fa.Channel <- ChannelData{
		IsAppending: isAppending,
		String:      s,
	}
}

func (fa *FileWriter) Close() {
	fa.Channel <- ChannelData{
		close: true,
	}
}

func fileChannel(fw *FileWriter, file *os.File) {
	defer fw.WaitGroup.Done()

	for chanData := range fw.Channel {
		if chanData.close {
			close(fw.Channel)
			break
		}

		if chanData.IsAppending {
			_, err := file.WriteString(chanData.String)
			if err != nil {
				panic(err)
			}
			continue
		}

		err := file.Truncate(0)
		if err != nil {
			panic(err)
		}

		_, err = file.Seek(0, 0)
		if err != nil {
			panic(err)
		}

		_, err = file.WriteString(chanData.String)
		if err != nil {
			panic(err)
		}
	}

	if file != nil {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}
}
