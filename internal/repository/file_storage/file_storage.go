package filestorage

import (
	"bufio"
	"os"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
)

type filestorage struct {
	path string
}

func MustGetFileStorage(path string) *filestorage {
	_, err := os.OpenFile(path, os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}
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
		logger.Log.Errorf("error opening file", "error", err)
		return []byte{}, nil
	}

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		logger.Log.Errorf("error scanning file", "error", err)
		return []byte{}, err
	}
	data := scanner.Bytes()
	return data, nil
}
