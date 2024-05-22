package writers

import (
	"os"
	"sync"
)

func NewFileWriter(file *os.File, manualOut bool) *FileWriter {
	ch := make(chan ChannelData, 1000)
	fw := &FileWriter{
		inChan:    ch,
		OutChan:   make(chan bool),
		manualOut: manualOut,
		file:      file,
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
	// 'true' on every completed file write
	OutChan   chan bool
	manualOut bool
	inChan    chan ChannelData
	file      *os.File
}

func (fa *FileWriter) WriteString(s string, isAppending bool) {
	fa.inChan <- ChannelData{
		IsAppending: isAppending,
		String:      s,
	}
}

func (fa *FileWriter) Close() {
	fa.inChan <- ChannelData{
		close: true,
	}
	fa.WaitGroup.Wait()
}

func fileChannel(fw *FileWriter, file *os.File) {
	defer fw.WaitGroup.Done()

	for chanData := range fw.inChan {
		if chanData.close {
			close(fw.inChan)
			close(fw.OutChan)
			break
		}

		if chanData.IsAppending {
			_, err := file.WriteString(chanData.String)
			if err != nil {
				panic(err)
			}
			if fw.manualOut {
				fw.OutChan <- true
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

		if fw.manualOut {
			fw.OutChan <- true
		}
	}

	if file != nil {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}
}
