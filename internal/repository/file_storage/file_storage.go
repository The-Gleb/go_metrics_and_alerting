package filestorage

import (
	"bufio"
	"os"
)

type filestorage struct {
	path string
}

func NewFileStorage(path string) *filestorage {
	return &filestorage{
		path: path,
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
