package app

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"sync/atomic"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/models"
)

var (
	ErrInvalidMetricType error = errors.New("invalid mertic type")
)

type Repositiries interface {
	// UpdateMetric(mType, mName, mValue string) error
	// GetMetric(mType, mName string) (string, error)
	GetAllMetrics() (*sync.Map, *sync.Map)
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (*float64, error)
	GetCounter(name string) (*int64, error)
}

type FileWriter interface {
	SaveMetrics(data []byte) error
	NeedToSyncWrite() bool
}

type app struct {
	storage         Repositiries
	fileStoragePath string
	storeInterval   int
}

func NewApp(s Repositiries, path string, interval int) *app {
	return &app{
		storage:         s,
		fileStoragePath: path,
		storeInterval:   interval,
	}
}

func (a *app) LoadDataFromFile() error {
	var maps models.MetricsMaps
	file, err := os.Open(a.fileStoragePath)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	data := scanner.Bytes()
	log.Printf("JSON data in file is %s\n\n", string(data))
	err = json.Unmarshal(data, &maps)
	if err != nil {
		logger.Log.Fatal(err)
	}
	// log.Printf("\ngauge map is %v\n", maps.Gauge)
	// log.Printf("\ncounter map is %v\n", maps.Counter)

	for k, v := range maps.Gauge {
		a.storage.UpdateGauge(k, v)
	}
	for k, v := range maps.Counter {
		a.storage.UpdateCounter(k, v)
	}
	stor, _ := a.GetAllMetricsJSON()
	log.Printf("loaded and got %v", string(stor))
	return nil
}

func (a *app) StoreDataToFile() {
	data, err := a.GetAllMetricsJSON()
	if err != nil {
		logger.Log.Fatal(err)
	}
	log.Printf("JSON data in file is %s", string(data))
	var maps models.MetricsMaps
	err = json.Unmarshal(data, &maps)
	if err != nil {
		logger.Log.Fatal(err)
	}
	log.Printf("\ngauge map is %v\n", maps.Gauge)
	log.Printf("\ncounter map is %v\n", maps.Counter)
	file, err := os.Create(a.fileStoragePath)
	if err != nil {
		log.Fatal("couldn`t open file to store data")
	}
	file.Write(data)
	file.Close()
}

func (a *app) UpdateMetricFromJSON(body io.Reader) ([]byte, error) {

	var metricObj models.Metrics
	var ret []byte

	err := json.NewDecoder(body).Decode(&metricObj)
	if err != nil {
		return ret, err
	}

	// log.Printf("struct is\n%v", metricObj)
	switch metricObj.MType {
	case "gauge":
		a.storage.UpdateGauge(metricObj.ID, *metricObj.Value)
		metricObj.Value, err = a.storage.GetGauge(metricObj.ID)
		if err != nil {
			return make([]byte, 0), err
		}
	case "counter":
		// log.Printf("\nPOST REQ BODY TO UPDATE %v", metricObj)
		// log.Printf("to update key: %s, val: %d", metricObj.ID, *metricObj.Delta)
		a.storage.UpdateCounter(metricObj.ID, *metricObj.Delta)
		metricObj.Delta, err = a.storage.GetCounter(metricObj.ID)
		if err != nil {
			return make([]byte, 0), err
		}
		// log.Printf("Updated key: %s, val: %d\n", metricObj.ID, *metricObj.Delta)
	default:
		return ret, ErrInvalidMetricType
	}
	if a.storeInterval == 0 {
		a.StoreDataToFile()
	}
	return json.Marshal(metricObj)
}

func (a *app) UpdateMetricFromParams(mType, mName, mValue string) ([]byte, error) {
	jsonObj, err := ParamsToJSON(mType, mName, mValue)

	if err != nil {
		return make([]byte, 0), err
	}
	return a.UpdateMetricFromJSON(bytes.NewBuffer(jsonObj))
}

func (a *app) GetMetricFromJSON(body io.Reader) ([]byte, error) {

	var metricObj models.Metrics
	var ret []byte
	err := json.NewDecoder(body).Decode(&metricObj)
	if err != nil {
		return ret, err
	}

	switch metricObj.MType {
	case "gauge":
		metricObj.Value, err = a.storage.GetGauge(metricObj.ID)
		if err != nil {
			return make([]byte, 0), err
		}
	case "counter":
		// log.Printf("\nREQ GET BOODY IS %v", metricObj)
		// log.Printf("Wanna get %s", metricObj.ID)
		metricObj.Delta, err = a.storage.GetCounter(metricObj.ID)
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

func (a *app) GetMetricFromParams(mType, mName string) ([]byte, error) {
	var strVal string
	switch mType {
	case "gauge":
		val, err := a.storage.GetGauge(mName)
		if err != nil {
			return make([]byte, 0), err
		}
		strVal = fmt.Sprintf("%v", *val)
	case "counter":
		val, err := a.storage.GetCounter(mName)
		if err != nil {
			return make([]byte, 0), err
		}
		strVal = fmt.Sprintf("%d", *val)
	default:
		return make([]byte, 0), ErrInvalidMetricType
	}
	return []byte(strVal), nil
}

func (a *app) GetAllMetricsJSON() ([]byte, error) {
	syncGaugeMap, syncCounterMap := a.storage.GetAllMetrics()
	maps := models.MetricsMaps{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}

	b := new(bytes.Buffer)

	syncGaugeMap.Range(func(key, value any) bool {
		maps.Gauge[key.(string)] = value.(float64)
		return true
	})
	syncCounterMap.Range(func(key, value any) bool {
		v := value.(*atomic.Int64).Load()

		maps.Counter[key.(string)] = v
		return true
	})

	jsonMaps, err := json.Marshal(&maps)
	// log.Printf("jsoned map %s\n", string(jsonMaps))
	if err != nil {
		return make([]byte, 0), err
	}

	fmt.Fprint(b, string(jsonMaps))
	return b.Bytes(), nil
}

func (a *app) GetAllMetricsHTML() []byte {
	gaugeMap, counterMap := a.storage.GetAllMetrics()

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
	gaugeMap.Range(func(key, value any) bool {
		fmt.Fprintf(b, "<li>%s = %f</li>", key, value)
		return true
	})
	counterMap.Range(func(key, value any) bool {
		fmt.Fprintf(b, "<li>%s = %d</li>", key, value)
		return true
	})
	fmt.Fprintf(b, "</ul></body></body>")
	return b.Bytes()
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
