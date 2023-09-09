package sink

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

type Storage interface {
	SinkJSON(filename string, data []byte) error
}

type jsonStorage struct {
	dir    string
	fs     afero.Fs
	dataCh chan *fileData
}

type fileData struct {
	filename string
	data     []byte
}

func NewJSONStorage(dir string) Storage {
	js := &jsonStorage{
		dir:    dir,
		fs:     afero.NewOsFs(),
		dataCh: make(chan *fileData, 1000), // buffer size, adjust as needed
	}

	go js.writeDataInBackground()

	return js
}

func (s *jsonStorage) SinkJSON(filename string, data []byte) error {
	s.dataCh <- &fileData{
		filename: filename,
		data:     data,
	}
	return nil
}

func (s *jsonStorage) writeDataInBackground() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	var buffer []*fileData

	for {
		select {
		case fd := <-s.dataCh:
			buffer = append(buffer, fd)
		case <-ticker.C:
			if len(buffer) == 0 {
				continue
			}

			filePath := filepath.Join(s.dir, buffer[0].filename)

			// Ensure directory exists before writing
			dir := filepath.Dir(filePath)
			if _, err := s.fs.Stat(dir); os.IsNotExist(err) {
				if err := s.fs.MkdirAll(dir, 0755); err != nil {
					// Log error or handle it in some other way
					continue
				}
			}

			file, err := s.fs.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

			if err != nil {
				// Log error and continue
				continue
			}
			defer file.Close()

			stat, err := file.Stat()
			if err != nil {
				// Log error and continue
				continue
			}

			if stat.Size() == 0 {
				file.Write([]byte("["))
			} else {
				file.Truncate(stat.Size() - 1) // remove the last character, which should be ']'
				file.Write([]byte(","))
			}

			for i, fd := range buffer {
				file.Write(fd.data)
				if i < len(buffer)-1 {
					file.Write([]byte(","))
				}
			}

			file.Write([]byte("]"))

			buffer = buffer[:0]
		}
	}
}
