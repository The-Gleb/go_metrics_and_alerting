package fileWriter

import (
	"io"
	"os"
	"time"
)

type App interface {
	UpdateMetricFromJSON(body io.Reader) ([]byte, error)
	UpdateMetricFromParams(mType, mName, mValue string) ([]byte, error)
	GetMetricFromParams(mType, mName string) ([]byte, error)
	GetMetricFromJSON(body io.Reader) ([]byte, error)
	GetAllMetricsHTML() []byte
	GetAllMetricsJSON() ([]byte, error)
	// ParamsToStruct(mType, mName, mValue string) (models.Metrics, error)
}

type fileWriter struct {
	restore  bool
	path     string
	interval time.Duration
}

func NewFileWriter(restore bool, path string, interval time.Duration) *fileWriter {
	// file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_SYNC, 0444)
	return &fileWriter{
		restore:  restore,
		path:     path,
		interval: interval,
	}
}

func (fw *fileWriter) NeedToSyncWrite() bool {
	return fw.interval == 0
}
func (fw *fileWriter) NeedToWrite() bool {
	return fw.interval > 0
}

func (fw *fileWriter) SaveMetrics(data []byte) error {

	file, err := os.Create(fw.path)
	if err != nil {
		return err
	}
	file.Write(data)
	file.Close()

	return nil
}
