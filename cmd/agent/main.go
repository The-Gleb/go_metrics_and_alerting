package main

import (
	// "compress/gzip"
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/models"
	"github.com/go-resty/resty/v2"
)

func main() {
	config := NewConfigFromFlags()

	logger.Initialize("debug")

	gaugeMap := make(map[string]float64)
	var PollCount int64 = 1

	var pollInterval = time.Duration(config.PollInterval) * time.Second
	var reportInterval = time.Duration(config.ReportInterval) * time.Second

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	baseURL := fmt.Sprintf("http://%s", config.Addres)
	client := resty.New()
	client.
		SetRetryCount(3).
		SetRetryMaxWaitTime(5 * time.Second).
		// SetRetryWaitTime(1 * time.Second).
		SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
			log.Printf("attempt: %d", r.Request.Attempt)
			dur := time.Duration(r.Request.Attempt*2-1) * time.Second
			return dur, nil
		}).
		SetBaseURL(baseURL)

	for {
		select {
		case <-pollTicker.C:
			CollectMetrics(gaugeMap, &PollCount)
		case <-reportTicker.C:
			// SendMetricsJSON(gaugeMap, &PollCount, req)
			SendMetricsInOneRequest(gaugeMap, &PollCount, client, []byte(config.SignKey))
		case <-c:
			pollTicker.Stop()
			reportTicker.Stop()

			return
		}
	}
}

func SendTestGet(req *resty.Request) {

	resp, _ := req.
		Get("/value/counter/PollCount")
	log.Println(string(resp.Body()))
	log.Println(resp.StatusCode())
}

func SendMetricsInOneRequest(gaugeMap map[string]float64, PollCount *int64, client *resty.Client, signKey []byte) {
	metrics := make([]models.Metrics, 0)

	for name, value := range gaugeMap {
		metrics = append(metrics, models.Metrics{
			MType: "gauge",
			ID:    name,
			Value: &value,
		})
	}
	metrics = append(metrics, models.Metrics{
		MType: "counter",
		ID:    "PollCount",
		Delta: PollCount,
	})

	data, err := json.Marshal(&metrics)
	if err != nil {
		log.Fatal(err)
	}

	logger.Log.Debug("sent body is", string(data))

	var sign []byte
	if len(signKey) > 0 {
		sign, err = hash(data, signKey)
		if err != nil {
			log.Fatal(err)
		}
		// sign, err = []byte(hex.EncodeToString())

		logger.Log.Debug("signKey is ", string(signKey))
		logger.Log.Debug("hex encoded signature is ", hex.EncodeToString(sign))
	}

	buf := bytes.Buffer{}
	gw := gzip.NewWriter(&buf)
	gw.Write(data)
	err = gw.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
	body := buf.Bytes()

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("HashSHA256", hex.EncodeToString(sign)).
		SetBody(body).
		Post("/updates/")
	if err != nil {
		return
	}
	log.Println(resp.Header().Get("Content-Encoding"))
	log.Println(string(resp.Body()))
}

func hash(data, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(data)
	if err != nil {
		return make([]byte, 0), err
	}

	sign := h.Sum(nil)

	return sign, nil
}

func SendMetricsJSON(gaugeMap map[string]float64, PollCount *int64, req *resty.Request) {
	for name, val := range gaugeMap {
		var result models.Metrics
		_, err := req.
			SetBody(&models.Metrics{
				ID:    name,
				MType: "gauge",
				Value: &val,
			}).
			SetResult(&result).
			Post("/update/")

		if err != nil {
			return
		}

	}
	var result models.Metrics
	_, err := req.
		SetBody(&models.Metrics{
			ID:    "PollCount",
			MType: "counter",
			Delta: PollCount,
		}).
		SetResult(&result).
		Post("/update/")

	if err != nil {
		return
	}
	log.Printf("\nUpdated to %v\n", result)
}

func SendMetrics(gaugeMap map[string]float64, PollCount *int64, client *resty.Client) {
	for name, val := range gaugeMap {
		requestURL := fmt.Sprintf("%s/update/gauge/%s/%f", client.BaseURL, name, val)

		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Accept-Encoding", "gzip").
			Post(requestURL)
		if err != nil {
			log.Printf("client: error making http request: %s\n", err)
			return
		}
		log.Println(string(resp.Body()))

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
	log.Printf("client: status code: %d\n", resp.StatusCode())
	log.Println(string(resp.Body()))

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
	// *counter++

	// b, _ := json.Marshal(gaugeMap)
	// fmt.Println(string(b))
	log.Printf("METRICS COLLECTED \n\n")

}

func SendTestGetJSON(req *resty.Request) {

	var result models.Metrics
	_, err := req.
		// SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		// SetHeader("Content-Encoding", "gzip").
		SetBody(&models.Metrics{
			ID:    "PollCount",
			MType: "counter",
		}).
		SetResult(&result).
		Post("/value/")

	if err != nil {
		return
	}
	log.Printf("PollCount is %d", *result.Delta)

}
