package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"
)

// TODO: create proper structure
// TODO: make unit tests
func main() {
	parseFlags()
	// map is not safe for concurrent use
	// TODO: implement concurrency-safe solution
	gaugeMap := make(map[string]float64)

	var PollCount int64 = 0
	var pollInterval = time.Duration(pollInterval) * time.Second
	var reportInterval = time.Duration(reportInterval) * time.Second

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	stop := make(chan bool)

	baseURL := fmt.Sprintf("http://%s", flagRunAddr)
	client := resty.New().
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetBaseURL(baseURL)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for {
			select {
			case <-pollTicker.C:
				CollectMetrics(gaugeMap, &PollCount)
			case <-stop:
				wg.Done()
				return
			}
		}
	}()
	wg.Add(1)
	go func() {
		for {
			select {
			case <-reportTicker.C:
				SendMetrics(gaugeMap, &PollCount, client)
			case <-stop:
				wg.Done()
				return
			}
		}
	}()

	// Блокировка, пока не будет получен сигнал
	<-c
	pollTicker.Stop()
	reportTicker.Stop()
	stop <- true
	stop <- true
	wg.Wait()
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

	// b, _ := json.Marshal(gaugeMap)
	// fmt.Println(string(b))
	log.Printf("METRICS COLLECTED \n\n")

}

func SendMetrics(gaugeMap map[string]float64, PollCount *int64, client *resty.Client) {
	for name, val := range gaugeMap {
		requestURL := fmt.Sprintf("%s/update/gauge/%s/%f", client.BaseURL, name, val)

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

	requestURL := fmt.Sprintf("%s/update/counter/PollCount/%d", client.BaseURL, *PollCount)
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		Post(requestURL)
	if err != nil {
		log.Printf("client: error making http request: %s\n", err)
		return
	}

	log.Printf("\n\nMETRICS WERE SENT TO THE SERVER!\n ADDRES: %s", client.BaseURL)
	log.Printf("client: status code: %d\n", res.StatusCode())
}
