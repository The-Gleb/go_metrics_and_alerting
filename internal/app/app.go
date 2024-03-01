package app

import (
	// "bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/models"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repositories"
	"github.com/The-Gleb/go_metrics_and_alerting/pkg/utils/retry"
)

var (
	ErrInvalidMetricType error = errors.New("invalid mertic type")
)

type FileStorage interface {
	WriteData(data []byte) error
	ReadData() ([]byte, error)
	SyncWrite() bool
}

type app struct {
	storage     repositories.Repositiries
	fileStorage FileStorage
}

// TODO: add FileWriter
func NewApp(s repositories.Repositiries, fs FileStorage) *app {
	return &app{
		storage:     s,
		fileStorage: fs,
	}
}

func (a *app) PingDB() error {
	return a.storage.PingDB()
}

func (a *app) LoadDataFromFile(ctx context.Context) error {

	data, err := a.fileStorage.ReadData()
	if err != nil {
		return fmt.Errorf("LoadDataFromFile: failed reading data: %w", err)
	}
	var maps models.MetricsMaps

	err = json.Unmarshal(data, &maps)
	if err != nil {
		return fmt.Errorf("LoadDataFromFile: failed unmarshalling: %w", err)
	}
	// log.Printf("\ngauge map is %v\n", maps.Gauge)
	// log.Printf("\ncounter map is %v\n", maps.Counter)

	for _, metric := range maps.Gauge {
		err := a.storage.UpdateGauge(ctx, metric)
		if err != nil {
			return fmt.Errorf("LoadDataFromFile: failded updating gauge: %w", err)
		}
	}
	for _, metric := range maps.Counter {
		err := a.storage.UpdateCounter(ctx, metric)
		if err != nil {
			return fmt.Errorf("LoadDataFromFile: failded updating counter: %w", err)
		}
	}

	return nil
}

func (a *app) StoreDataToFile(ctx context.Context) error {
	data, err := a.GetAllMetricsJSON(ctx)
	if err != nil {
		return fmt.Errorf("StoreDataToFile: %w", err)
	}
	err = a.fileStorage.WriteData(data)
	if err != nil {
		return fmt.Errorf("StoreDataToFile: %w", err)
	}
	return nil
}

func (a *app) UpdateMetricSet(ctx context.Context, body io.Reader) ([]byte, error) {

	metrics := make([]models.Metrics, 0)

	err := json.NewDecoder(body).Decode(&metrics)
	if err != nil {
		return []byte{}, fmt.Errorf("UpdateMetricSet: %w", err)
	}

	if len(metrics) == 0 {
		return []byte{}, fmt.Errorf("UpdateMetricSet: %w", fmt.Errorf("no metrics were sent"))
	}

	var n int64
	err = retry.DefaultRetry(ctx, func(ctx context.Context) error {
		n, err = a.storage.UpdateMetricSet(ctx, metrics)
		return err
	})

	// n, err := a.storage.UpdateMetricSet(ctx, metrics)
	if err != nil {
		return []byte{}, fmt.Errorf("UpdateMetricSet: %w", err)
	}

	b := bytes.Buffer{}
	b.WriteString(fmt.Sprint(n))
	b.WriteString(" metrics were successfuly updated")
	// ret := fmt.Sprintf("%d metrics were successfuly updated", n)

	return b.Bytes(), nil
}

func (a *app) UpdateMetricFromJSON(ctx context.Context, body io.Reader) ([]byte, error) {

	var metricObj models.Metrics

	err := json.NewDecoder(body).Decode(&metricObj)
	if err != nil {
		return []byte{}, fmt.Errorf("UpdateMetricFromJSON: %w", err)
	}

	switch metricObj.MType {
	case "gauge":
		err = retry.DefaultRetry(context.TODO(), func(ctx context.Context) error {
			err = a.storage.UpdateGauge(ctx, metricObj)
			return err
		})

		if err != nil {
			return []byte{}, fmt.Errorf("UpdateMetricFromJSON: %w", err)
		}
		metricObj, err = a.storage.GetGauge(ctx, metricObj)
		if err != nil {
			return []byte{}, fmt.Errorf("UpdateMetricFromJSON: %w", err)
		}

	case "counter":
		err = retry.DefaultRetry(context.TODO(), func(ctx context.Context) error {
			err = a.storage.UpdateCounter(ctx, metricObj)
			return err
		})

		if err != nil {
			return []byte{}, fmt.Errorf("UpdateMetricFromJSON: %w", err)
		}
		metricObj, err = a.storage.GetCounter(ctx, metricObj)
		if err != nil {
			return []byte{}, fmt.Errorf("UpdateMetricFromJSON: %w", err)
		}

	default:
		return []byte{}, ErrInvalidMetricType
	}
	if a.fileStorage != nil && a.fileStorage.SyncWrite() {
		a.StoreDataToFile(ctx)
	}
	return json.Marshal(metricObj)
}

func (a *app) UpdateMetricFromParams(ctx context.Context, mType, mName, mValue string) ([]byte, error) {
	jsonObj, err := ParamsToJSON(mType, mName, mValue)

	if err != nil {
		return []byte{}, fmt.Errorf("UpdateMetricFromParams: %w", err)
	}
	return a.UpdateMetricFromJSON(ctx, bytes.NewBuffer(jsonObj))
}

func (a *app) GetMetricFromJSON(ctx context.Context, body io.Reader) ([]byte, error) {

	var metricObj models.Metrics

	err := json.NewDecoder(body).Decode(&metricObj)
	if err != nil {
		return []byte{}, fmt.Errorf("GetMetricFromJSON: %w", err)
	}
	logger.Log.Debugw("metricObj is", "struct", metricObj)
	switch metricObj.MType {
	case "gauge":
		err = retry.DefaultRetry(context.TODO(), func(ctx context.Context) error {
			metricObj, err = a.storage.GetGauge(ctx, metricObj)
			return err
		})
		// metricObj, err = a.storage.GetGauge(ctx, metricObj)
		if err != nil {
			return []byte{}, fmt.Errorf("GetMetricFromJSON: %w", err)
		}
	case "counter":
		err = retry.DefaultRetry(context.TODO(), func(ctx context.Context) error {
			metricObj, err = a.storage.GetCounter(ctx, metricObj)
			return err
		})
		// metricObj, err = a.storage.GetCounter(ctx, metricObj)
		if err != nil {
			return []byte{}, fmt.Errorf("GetMetricFromJSON: %w", err)
		}

	default:
		return []byte{}, ErrInvalidMetricType
	}
	return json.Marshal(metricObj)
}

func (a *app) GetMetricFromParams(ctx context.Context, mType, mName string) ([]byte, error) {
	jsonObj, err := ParamsToJSON(mType, mName, "")

	if err != nil {
		return []byte{}, fmt.Errorf("GetMetricFromParams: %w", err)
	}
	data, err := a.GetMetricFromJSON(ctx, bytes.NewBuffer(jsonObj))
	if err != nil {
		return []byte{}, fmt.Errorf("GetMetricFromParams: %w", err)
	}
	var metricObj models.Metrics
	err = json.Unmarshal(data, &metricObj)
	if err != nil {
		return []byte{}, fmt.Errorf("GetMetricFromParams: %w", err)
	}
	b := new(bytes.Buffer)
	switch metricObj.MType {
	case "gauge":
		_, err = fmt.Fprintf(b, "%v", *metricObj.Value)
		if err != nil {
			return []byte{}, fmt.Errorf("GetMetricFromParams: %w", err)
		}
		return b.Bytes(), err
	case "counter":
		_, err = fmt.Fprintf(b, "%v", *metricObj.Delta)
		if err != nil {
			return []byte{}, fmt.Errorf("GetMetricFromParams: %w", err)
		}
		return b.Bytes(), err
	default:
		return b.Bytes(), ErrInvalidMetricType
	}
}

func (a *app) GetAllMetricsJSON(ctx context.Context) ([]byte, error) {
	var gaugeMap, counterMap []models.Metrics
	var err error
	err = retry.DefaultRetry(context.TODO(), func(ctx context.Context) error {
		gaugeMap, counterMap, err = a.storage.GetAllMetrics(ctx)
		return err
	})
	if err != nil {
		return []byte{}, fmt.Errorf("GetAllMetricsJSON: %w", err)
	}
	maps := models.MetricsMaps{
		Gauge:   gaugeMap,
		Counter: counterMap,
	}

	b := new(bytes.Buffer)

	jsonMaps, err := json.Marshal(&maps)
	if err != nil {
		return []byte{}, fmt.Errorf("GetAllMetricsJSON: %w", err)
	}

	fmt.Fprint(b, string(jsonMaps))
	return b.Bytes(), nil
}

func (a *app) GetAllMetricsHTML(ctx context.Context) ([]byte, error) {
	var gaugeMap, counterMap []models.Metrics
	var err error
	err = retry.DefaultRetry(context.TODO(), func(ctx context.Context) error {
		gaugeMap, counterMap, err = a.storage.GetAllMetrics(ctx)
		return err
	})
	if err != nil {
		return []byte{}, fmt.Errorf("GetAllMetricsHTML: %w", err)
	}
	b := new(bytes.Buffer)
	fmt.Fprintf(b, `
	<html>
		<head>
			<meta charset="utf-8">
			<title>list-style-type</title>
			<style>
				ul {
					list-style-type: none;
				}
			</style>
		</head>
		<body>
		<ul>`)
	for _, metric := range gaugeMap {
		fmt.Fprintf(b, "<li>%s = %f</li>", metric.ID, *metric.Value)
	}
	for _, metric := range counterMap {
		fmt.Fprintf(b, "<li>%s = %d</li>", metric.ID, *metric.Delta)
	}
	fmt.Fprintf(b, "</ul></body></body>")
	return b.Bytes(), nil
}

func ParamsToJSON(mType, mName, mValue string) ([]byte, error) {
	b := new(bytes.Buffer)
	// var json string
	switch mType {
	case "gauge":

		_, err := fmt.Fprintf(b, `{
			"id": "%s",
			"type": "%s"`, mName, mType)

		if err != nil {
			return []byte{}, fmt.Errorf("ParamsToJSON: %w", err)
		}

		if mValue != "" {
			_, err = fmt.Fprintf(b, `,"value": %s
			}`, mValue)
		} else {
			_, err = fmt.Fprintf(b, `}`)
		}

		if err != nil {
			return []byte{}, fmt.Errorf("ParamsToJSON: %w", err)
		}
	case "counter":
		_, err := fmt.Fprintf(b, `{
			"id": "%s",
			"type": "%s"`, mName, mType)

		if err != nil {
			return []byte{}, fmt.Errorf("ParamsToJSON: %w", err)
		}
		if mValue != "" {
			_, err = fmt.Fprintf(b, `,"delta": %s
			}`, mValue)
		} else {
			_, err = fmt.Fprintf(b, `}`)
		}

		if err != nil {
			return []byte{}, fmt.Errorf("ParamsToJSON: %w", err)
		}
	default:
		return b.Bytes(), ErrInvalidMetricType
	}

	return b.Bytes(), nil
}
