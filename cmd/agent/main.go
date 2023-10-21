package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/go-resty/resty/v2"
)

func main() {
	config := NewConfigFromFlags()

	gaugeMap := make(map[string]float64)

	var PollCount int64 = 0
	var pollInterval = time.Duration(config.PollInterval) * time.Second
	var reportInterval = time.Duration(config.ReportInterval) * time.Second

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	// stop := make(chan bool)

	baseURL := fmt.Sprintf("http://%s", config.Addres)
	client := resty.New().
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetBaseURL(baseURL)

	var wg sync.WaitGroup
	wg.Add(1)

	for {
		select {
		case <-pollTicker.C:
			CollectMetrics(gaugeMap, &PollCount)
		case <-reportTicker.C:
			SendMetrics(gaugeMap, &PollCount, client)
		case <-c:
			pollTicker.Stop()
			reportTicker.Stop()
			return
		}
	}
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

		_, err := client.R().
			SetHeader("Content-Type", "application/json").
			Post(requestURL)
		if err != nil {
			log.Printf("client: error making http request: %s\n", err)
			return
		}

	}

	requestURL := fmt.Sprintf("%s/update/counter/PollCount/%d", client.BaseURL, *PollCount)
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		Post(requestURL)
	if err != nil {
		log.Printf("client: error making http request: %s\n", err)
		return
	}

	logger.Log.Infow("METRICS SENT - : %s\nStatus: %d\n",
		"ADDRES", client.BaseURL,
		"Status", resp.StatusCode(),
	)
	// log.Printf("client: status code: %d\n", res.StatusCode())

}
