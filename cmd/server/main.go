package main

import (
	// "fmt"
	"encoding/json"
	"net/http"
	"strconv"

	//"runtime/metrics"
	"strings"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

type MemStorage struct {
	gauge   map[string]float64
	counter map[string][]int64
}

var memStorage MemStorage = MemStorage{
	gauge:   make(map[string]float64),
	counter: make(map[string][]int64),
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", UpdateMetric)
	return http.ListenAndServe(`:8080`, mux)
}

func UpdateMetric(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rw.Header().Add("Content-Type", "text/plain; charset=UTF-8")
	pathValues := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")

	switch len(pathValues) {
	case 2:
		http.Error(rw, "metric value was not sent", http.StatusBadRequest)
		return
	case 1:
		http.Error(rw, "metric name was not sent", http.StatusNotFound)
		return
	case 0:
		http.Error(rw, "metric type was not sent", http.StatusBadRequest)
		return
	}

	m_Type := pathValues[0]
	m_Name := pathValues[1]

	switch m_Type {
	case "gauge":
		// _, ok := pathValues[2]
		m_Value, err := strconv.ParseFloat(pathValues[2], 64)
		if err != nil {
			http.Error(rw, "incorrect metric value\ncannot parse to float64", http.StatusBadRequest)
			return
		}
		memStorage.gauge[m_Name] = m_Value
		// body, err := json.Marshal(memStorage.gauge)
		// if err != nil {

		// }
		// rw.Write(body)
	case "counter":
		m_Value, err := strconv.ParseInt(pathValues[2], 10, 32)
		if err != nil {
			http.Error(rw, "incorrect metric value\ncannot parse to int32", http.StatusBadRequest)
			return
		}
		memStorage.counter[m_Name] = append(memStorage.counter[m_Name], m_Value)
		// body, err := json.Marshal(memStorage.counter)
		// if err != nil {

		// }
		// rw.Write(body)
	default:
		http.Error(rw, "Metric not found", http.StatusBadRequest)
	}
	rw.WriteHeader(http.StatusOK)

}
