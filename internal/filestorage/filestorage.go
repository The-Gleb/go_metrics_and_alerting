package filestorage

import (
	"bufio"
	"os"
)

type filestorage struct {
	path          string
	storeInterval int
	restore       bool
}

type FileWriter interface {
	WriteData(data []byte) error
	ReadData() ([]byte, error)
	SyncWrite() bool
}

func NewFileStorage(path string, interval int, restore bool) *filestorage {
	return &filestorage{
		path:          path,
		storeInterval: interval,
		restore:       restore,
	}
}

func (fs *filestorage) WriteData(data []byte) error {
	file, err := os.Create(fs.path)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

func (fs *filestorage) ReadData() ([]byte, error) {
	file, err := os.Open(fs.path)
	if err != nil {
		return make([]byte, 0), nil
	}

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return make([]byte, 0), err
	}
	data := scanner.Bytes()
	return data, nil
}

func (fs *filestorage) SyncWrite() bool {
	return fs.storeInterval == 0
}
