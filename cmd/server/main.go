package main

import (
	"fmt"
	"log"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/handlers"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/server"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/storage"
)

func main() {
	parseFlags()
	storage := storage.New()
	handlers := handlers.New(storage)
	baseURL := fmt.Sprintf("http://%s", flagRunAddr)

	server := server.New(flagRunAddr, handlers)
	log.Println(flagRunAddr)
	log.Println(baseURL)

	err := server.Run()
	if err != nil {
		panic(err)
	}
	log.Printf("server started")
}

// func main() {
// 	if err := run(); err != nil {
// 		panic(err)
// 	}
// }

// type MemStorage struct {
// 	gauge   map[string]float64
// 	counter map[string][]int64
// }

// var memStorage MemStorage = MemStorage{
// 	gauge:   make(map[string]float64),
// 	counter: make(map[string][]int64),
// }

// func run() error {
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/update/", UpdateMetric)
// 	return http.ListenAndServe(`:8080`, mux)
// }

// func UpdateMetric(rw http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		rw.WriteHeader(http.StatusMethodNotAllowed)
// 		return
// 	}
// 	rw.Header().Add("Content-Type", "text/plain; charset=UTF-8")
// 	pathValues := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")

// 	switch len(pathValues) {
// 	case 2:
// 		http.Error(rw, "metric value was not sent", http.StatusNotFound)
// 		return
// 	case 1:
// 		http.Error(rw, "metric name was not sent", http.StatusNotFound)
// 		return
// 	case 0:
// 		http.Error(rw, "metric type was not sent", http.StatusBadRequest)
// 		return
// 	}

// 	mType := pathValues[0]
// 	mName := pathValues[1]

// 	switch mType {
// 	case "gauge":
// 		// _, ok := pathValues[2]
// 		mValue, err := strconv.ParseFloat(pathValues[2], 64)
// 		if err != nil {
// 			http.Error(rw, "incorrect metric value\ncannot parse to float64", http.StatusBadRequest)
// 			return
// 		}
// 		memStorage.gauge[mName] = mValue
// 		// body, err := json.Marshal(memStorage.gauge)
// 		// if err != nil {

// 		// }
// 		// rw.Write(body)
// 	case "counter":
// 		mValue, err := strconv.ParseInt(pathValues[2], 10, 32)
// 		if err != nil {
// 			http.Error(rw, "incorrect metric value\ncannot parse to int32", http.StatusBadRequest)
// 			return
// 		}
// 		memStorage.counter[mName] = append(memStorage.counter[mName], mValue)
// 		// body, err := json.Marshal(memStorage.counter)
// 		// if err != nil {

// 		// }
// 		// rw.Write(body)
// 	default:
// 		http.Error(rw, "Metric not found", http.StatusBadRequest)
// 	}
// 	rw.WriteHeader(http.StatusOK)

// }
