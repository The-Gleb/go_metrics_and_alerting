package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"

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

type app struct {
	storage Repositiries
}

func NewApp(s Repositiries) *app {
	return &app{
		storage: s,
	}
}

func (a *app) UpdateMetricFromJSON(body io.Reader) ([]byte, error) {
	var metricObj models.Metrics
	var ret []byte
	err := json.NewDecoder(body).Decode(&metricObj)
	if err != nil {
		return ret, err
	}
	switch metricObj.ID {
	case "gauge":
		a.storage.UpdateGauge(metricObj.MType, *metricObj.Value)
		metricObj.Value, err = a.storage.GetGauge(metricObj.MType)
		if err != nil {
			return make([]byte, 0), err
		}
	case "counter":
		a.storage.UpdateCounter(metricObj.MType, *metricObj.Delta)
		metricObj.Delta, err = a.storage.GetCounter(metricObj.MType)
		if err != nil {
			return make([]byte, 0), err
		}
	default:
		return ret, ErrInvalidMetricType
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
	// log.Printf("Struct is: \n%v\n", metricObj)

	switch metricObj.ID {
	case "gauge":
		metricObj.Value, err = a.storage.GetGauge(metricObj.MType)
		if err != nil {
			return make([]byte, 0), err
		}
	case "counter":
		metricObj.Delta, err = a.storage.GetCounter(metricObj.MType)
		if err != nil {
			return make([]byte, 0), err
		}
		log.Printf("Changed struct is: \n%v\n", metricObj)

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
	gaugeMap, counterMap := a.storage.GetAllMetrics()
	b := new(bytes.Buffer)
	fmt.Fprintf(b, "{")

	gaugeMap.Range(func(key, value any) bool {
		fmt.Fprintf(b, `"%s":%v,`, key, value)
		return true
	})
	counterMap.Range(func(key, value any) bool {
		fmt.Fprintf(b, `"%s":%v`, key, value)
		return true
	})

	fmt.Fprintf(b, "}")
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
		}`, mType, mName, mValue)
	case "counter":
		json = fmt.Sprintf(`{
			"id": "%s",
			"type": "%s",
			"delta": %s
		}`, mType, mName, mValue)
	default:
		return make([]byte, 0), ErrInvalidMetricType
	}

	return []byte(json), nil
}
