package handlers

import (
	// "fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/storage"
)

type handlers struct {
	storage storage.Repositiries
}

type Handlers interface {
	UpdateMetric(rw http.ResponseWriter, r *http.Request)
}

func New(store storage.Repositiries) *handlers {

	return &handlers{
		storage: store,
	}
}

func (handlers *handlers) UpdateMetric(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rw.Header().Add("Content-Type", "text/plain; charset=UTF-8")
	pathValues := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")

	switch len(pathValues) {
	case 2:
		http.Error(rw, "metric value was not sent", http.StatusNotFound)
		return
	case 1:
		http.Error(rw, "metric name was not sent", http.StatusNotFound)
		return
	case 0:
		http.Error(rw, "metric type was not sent", http.StatusBadRequest)
		return
	}

	mType := pathValues[0]
	mName := pathValues[1]

	switch mType {
	case "gauge":
		mValue, err := strconv.ParseFloat(pathValues[2], 64)
		if err != nil {
			http.Error(rw, "incorrect metric value\ncannot parse to float64", http.StatusBadRequest)
			return
		}
		handlers.storage.UpdateGauge(mName, mValue)
	case "counter":
		mValue, err := strconv.ParseInt(pathValues[2], 10, 32)
		if err != nil {
			http.Error(rw, "incorrect metric value\ncannot parse to int32", http.StatusBadRequest)
			return
		}
		handlers.storage.UpdateCounter(mName, mValue)
	default:
		http.Error(rw, "Invalid mertic type", http.StatusBadRequest)
	}
	rw.WriteHeader(http.StatusOK)
	// fmt.Println("request was processed successfuly")

}
