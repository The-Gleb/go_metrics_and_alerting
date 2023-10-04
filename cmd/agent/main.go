package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"math/rand"
	// "net/http"
	"os"
	"os/signal"
	"runtime"
	"time"
)

func main() {
	parseFlags()
	gaugeMap := make(map[string]float64)
	var PollCount int64 = 0
	var pollInterval = time.Duration(pollInterval) * time.Second
	var reportInterval = time.Duration(reportInterval) * time.Second

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	updTicker := time.NewTicker(pollInterval)
	sendTicker := time.NewTicker(reportInterval)

	stop := make(chan bool)

	client := resty.New().
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second)

	go func() {
		defer func() { stop <- true }()
		for {
			select {
			case <-updTicker.C:
				CollectMetrics(gaugeMap, &PollCount)
			case <-stop:

				// fmt.Printf("Закрытие горутины %v\n", n)
				return
			}
		}
	}()

	go func() {
		defer func() { stop <- true }()
		for {
			select {
			case <-sendTicker.C:
				SendMetrics(gaugeMap, &PollCount, client)
			case <-stop:
				// fmt.Printf("Закрытие горутины %v\n", n)
				return
			}
		}
	}()

	// Блокировка, пока не будет получен сигнал
	test := <-c
	fmt.Println("stop signal", test)
	updTicker.Stop()
	sendTicker.Stop()

	// Остановка горутины
	stop <- true

	// Ожидание до тех пор, пока не выполнится
	<-stop
}

func CollectMetrics(gaugeMap map[string]float64, counter *int64) {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	gaugeMap["Alloc"] = float64(rtm.Alloc)
	gaugeMap["BuckHashSys"] = float64(rtm.BuckHashSys)
	gaugeMap["Frees"] = float64(rtm.Frees)
	gaugeMap["GCCPUFraction"] = float64(rtm.GCCPUFraction)
	gaugeMap["GCSys"] = float64(rtm.GCSys)
	gaugeMap["HeapAlloc"] = float64(rtm.HeapAlloc)
	gaugeMap["HeapIdle"] = float64(rtm.HeapIdle)
	gaugeMap["HeapInuse"] = float64(rtm.HeapInuse)
	gaugeMap["HeapObjects"] = float64(rtm.HeapObjects)
	gaugeMap["HeapReleased"] = float64(rtm.HeapReleased)
	gaugeMap["HeapSys"] = float64(rtm.HeapSys)
	gaugeMap["LastGC"] = float64(rtm.LastGC)
	gaugeMap["Lookups"] = float64(rtm.Lookups)
	gaugeMap["MCacheInuse"] = float64(rtm.MCacheInuse)
	gaugeMap["MCacheSys"] = float64(rtm.MCacheSys)
	gaugeMap["MSpanInuse"] = float64(rtm.MSpanInuse)
	gaugeMap["MSpanSys"] = float64(rtm.MSpanSys)
	gaugeMap["Mallocs"] = float64(rtm.Mallocs)
	gaugeMap["NextGC"] = float64(rtm.NextGC)
	gaugeMap["NumForcedGC"] = float64(rtm.NumForcedGC)
	gaugeMap["NumGC"] = float64(rtm.NumGC)
	gaugeMap["OtherSys"] = float64(rtm.OtherSys)
	gaugeMap["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	gaugeMap["StackInuse"] = float64(rtm.StackInuse)
	gaugeMap["StackSys"] = float64(rtm.StackSys)
	gaugeMap["Sys"] = float64(rtm.Sys)
	gaugeMap["TotalAlloc"] = float64(rtm.TotalAlloc)
	gaugeMap["RandomValue"] = rand.Float64()
	*counter++

	// Just encode to json and print
	// b, _ := json.Marshal(gaugeMap)
	// fmt.Println(string(b))
	log.Printf("METRICS COLLECTED \n\n")

}

func SendMetrics(gaugeMap map[string]float64, PollCount *int64, client *resty.Client) {
	for name, val := range gaugeMap {
		requestURL := fmt.Sprintf("http://127.0.0.1:8080/update/gauge/%s/%f", name, val)

		res, err := client.R().
			SetHeader("Content-Type", "application/json").
			Post(requestURL)
		if err != nil {
			log.Printf("client: error making http request: %s\n", err)
			return
		}

		log.Printf("client: status code: %d ", res.StatusCode())
	}
	log.Printf("gauge metrics sent")

	requestURL := fmt.Sprintf("http://127.0.0.1:8080/update/counter/PollCount/%d", *PollCount)
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		Post(requestURL)
	if err != nil {
		log.Printf("client: error making http request: %s\n", err)
		return
	}

	log.Printf("\n\nMETRICS WERE SENT TO THE SERVER!\n")
	log.Printf("client: status code: %d\n", res.StatusCode())
}

func MakeHTTPCall(URL string) {

}
