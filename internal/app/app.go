package app

import (
	// "bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/models"
)

var (
	ErrInvalidMetricType error = errors.New("invalid mertic type")
)

type Repositiries interface {
	GetAllMetrics(ctx context.Context) ([]models.Metrics, []models.Metrics, error)
	UpdateGauge(ctx context.Context, metricObj models.Metrics) error
	UpdateCounter(ctx context.Context, metricObj models.Metrics) error
	UpdateMetricSet(ctx context.Context, metrics []models.Metrics) (int64, error)
	GetGauge(ctx context.Context, metricObj models.Metrics) (models.Metrics, error)
	GetCounter(ctx context.Context, metricObj models.Metrics) (models.Metrics, error)
	PingDB() error
}

type FileStorage interface {
	WriteData(data []byte) error
	ReadData() ([]byte, error)
	SyncWrite() bool
}

type app struct {
	storage     Repositiries
	fileStorage FileStorage
}

// TODO: add FileWriter
func NewApp(s Repositiries, fs FileStorage) *app {
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
		return err
	}
	var maps models.MetricsMaps

	log.Printf("JSON data in file is %s\n\n", string(data))
	err = json.Unmarshal(data, &maps)
	if err != nil {
		return err
	}
	// log.Printf("\ngauge map is %v\n", maps.Gauge)
	// log.Printf("\ncounter map is %v\n", maps.Counter)

	for _, metric := range maps.Gauge {
		err := a.storage.UpdateGauge(ctx, metric)
		if err != nil {
			return err
		}
	}
	for _, metric := range maps.Counter {
		err := a.storage.UpdateCounter(ctx, metric)
		if err != nil {
			return err
		}
	}

	// just check
	stor, _ := a.GetAllMetricsJSON(ctx)
	log.Printf("loaded and got %v", string(stor))

	return nil
}

func (a *app) StoreDataToFile(ctx context.Context) error {
	data, err := a.GetAllMetricsJSON(ctx)
	if err != nil {
		return err
	}
	err = a.fileStorage.WriteData(data)
	if err != nil {
		return err
	}
	return nil
}

func (a *app) UpdateMetricSet(ctx context.Context, body io.Reader) ([]byte, error) {
	metrics := make([]models.Metrics, 0)
	err := json.NewDecoder(body).Decode(&metrics)
	if err != nil {
		return make([]byte, 0), err
	}
	if len(metrics) == 0 {
		return make([]byte, 0), fmt.Errorf("no metrics were sent")
	}
	n, err := a.storage.UpdateMetricSet(ctx, metrics)
	if err != nil {
		return make([]byte, 0), err
	}
	ret := fmt.Sprintf("%d metrics were successfuly updated", n)
	return []byte(ret), nil
}

func (a *app) UpdateMetricFromJSON(ctx context.Context, body io.Reader) ([]byte, error) {

	var metricObj models.Metrics
	var ret []byte

	err := json.NewDecoder(body).Decode(&metricObj)
	if err != nil {
		return ret, err
	}

	log.Printf("struct is\n%v", metricObj)
	switch metricObj.MType {
	case "gauge":
		err := a.storage.UpdateGauge(ctx, metricObj)
		if err != nil {
			return make([]byte, 0), err
		}
		metricObj, err = a.storage.GetGauge(ctx, metricObj)
		if err != nil {
			return make([]byte, 0), err
		}
	case "counter":
		// log.Printf("\nPOST REQ BODY TO UPDATE %v", metricObj)
		// log.Printf("to update key: %s, val: %d", metricObj.ID, *metricObj.Delta)
		err := a.storage.UpdateCounter(ctx, metricObj)
		if err != nil {
			return make([]byte, 0), err
		}
		metricObj, err = a.storage.GetCounter(ctx, metricObj)
		if err != nil {
			return make([]byte, 0), err
		}
		// log.Printf("Updated key: %s, val: %d\n", metricObj.ID, *metricObj.Delta)
	default:
		return ret, ErrInvalidMetricType
	}
	if a.fileStorage != nil && a.fileStorage.SyncWrite() {
		a.StoreDataToFile(ctx)
	}
	return json.Marshal(metricObj)
}

func (a *app) UpdateMetricFromParams(ctx context.Context, mType, mName, mValue string) ([]byte, error) {
	jsonObj, err := ParamsToJSON(mType, mName, mValue)

	if err != nil {
		return make([]byte, 0), err
	}
	return a.UpdateMetricFromJSON(ctx, bytes.NewBuffer(jsonObj))
}

func (a *app) GetMetricFromJSON(ctx context.Context, body io.Reader) ([]byte, error) {

	var metricObj models.Metrics
	var ret []byte
	err := json.NewDecoder(body).Decode(&metricObj)
	if err != nil {
		return ret, err
	}

	switch metricObj.MType {
	case "gauge":
		metricObj, err = a.storage.GetGauge(ctx, metricObj)
		if err != nil {
			return make([]byte, 0), err
		}
	case "counter":
		// log.Printf("\nREQ GET BOODY IS %v", metricObj)
		// log.Printf("Wanna get %s", metricObj.ID)
		metricObj, err = a.storage.GetCounter(ctx, metricObj)
		if err != nil {
			return make([]byte, 0), err
		}
		// log.Printf("Got err %v ", err)
		// log.Printf("\n GoT BOODY %v", metricObj)
		// log.Printf("Got value %d\n", *metricObj.Delta)

	default:
		return ret, ErrInvalidMetricType
	}
	return json.Marshal(metricObj)
}

func (a *app) GetMetricFromParams(ctx context.Context, mType, mName string) ([]byte, error) {
	var result string
	metricObj := models.Metrics{ID: mName, MType: mType}
	switch mType {
	case "gauge":
		metricObj, err := a.storage.GetGauge(ctx, metricObj)
		if err != nil {
			return make([]byte, 0), err
		}
		result = fmt.Sprintf("%v", *metricObj.Value)
	case "counter":
		metricObj, err := a.storage.GetCounter(ctx, metricObj)
		if err != nil {
			return make([]byte, 0), err
		}
		result = fmt.Sprintf("%d", *metricObj.Delta)
	default:
		return make([]byte, 0), ErrInvalidMetricType
	}
	return []byte(result), nil
}

func (a *app) GetAllMetricsJSON(ctx context.Context) ([]byte, error) {
	gaugeMap, counterMap, err := a.storage.GetAllMetrics(ctx)
	if err != nil {
		return make([]byte, 0), err
	}
	maps := models.MetricsMaps{
		Gauge:   gaugeMap,
		Counter: counterMap,
	}

	b := new(bytes.Buffer)

	jsonMaps, err := json.Marshal(&maps)
	// log.Printf("jsoned map %s\n", string(jsonMaps))
	if err != nil {
		return make([]byte, 0), err
	}

	fmt.Fprint(b, string(jsonMaps))
	return b.Bytes(), nil
}

func (a *app) GetAllMetricsHTML(ctx context.Context) ([]byte, error) {
	gaugeMap, counterMap, err := a.storage.GetAllMetrics(ctx)
	if err != nil {
		return make([]byte, 0), err
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
	var json string
	switch mType {
	case "gauge":
		json = fmt.Sprintf(`{
			"id": "%s",
			"type": "%s",
			"value": %s
		}`, mName, mType, mValue)
	case "counter":
		json = fmt.Sprintf(`{
			"id": "%s",
			"type": "%s",
			"delta": %s
		}`, mName, mType, mValue)
	default:
		return make([]byte, 0), ErrInvalidMetricType
	}

	return []byte(json), nil
}
