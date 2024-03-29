package main

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	cryptoRand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"net/http"
	_ "net/http/pprof"

	"github.com/go-resty/resty/v2"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
)

var (
	BuildVersion string = "N/A"
	BuildDate    string = "N/A"
	BuildCommit  string = "N/A"
)

func main() {
	fmt.Printf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		BuildVersion, BuildDate, BuildCommit,
	)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	config, err := BuildConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger.Initialize(config.LogLevel)

	logger.Log.Info(config)

	gaugeMap := make(map[string]float64)
	var pollCount atomic.Int64
	pollCount.Store(1)

	var pollInterval = time.Duration(config.PollInterval * 1000000000)
	var reportInterval = time.Duration(config.ReportInterval * 1000000000)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	baseURL := fmt.Sprintf("http://%s", config.Address)
	client := resty.New()
	client.
		SetRetryCount(3).
		SetRetryMaxWaitTime(5 * time.Second).
		SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
			logger.Log.Debug("attempt: %d", r.Request.Attempt)
			dur := time.Duration(r.Request.Attempt*2-1) * time.Second
			return dur, nil
		}).
		SetBaseURL(baseURL)

	sendTaskCh := make(chan struct{}, 1)
	collectTaskCh := make(chan struct{}, 1)
	collectMemsTaskCh := make(chan struct{}, 1)
	var mu sync.RWMutex

	for i := 0; i < config.RateLimit; i++ {
		go func() {
			for range sendTaskCh {
				SendMetricSet(gaugeMap, &pollCount, client, []byte(config.SignKey), config.PublicKeyPath, &mu)
			}
		}()
	}

	go func() {
		for range collectTaskCh {
			CollectMetrics(gaugeMap, &mu)
		}
	}()

	go func() {
		for range collectMemsTaskCh {
			CollectMemMetrics(gaugeMap, &mu)
		}
	}()

	for {
		select {
		case <-pollTicker.C:
			collectTaskCh <- struct{}{}
			collectMemsTaskCh <- struct{}{}
		case <-reportTicker.C:
			sendTaskCh <- struct{}{}
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
	logger.Log.Debug(string(resp.Body()))
	logger.Log.Debug(resp.StatusCode())
}

func SendMetricSet(
	gaugeMap map[string]float64, pollCount *atomic.Int64,
	client *resty.Client, signKey []byte, publicKeyPath string,
	mu *sync.RWMutex,
) {
	metrics := make([]entity.Metric, 0)

	mu.RLock()
	for name, value := range gaugeMap {
		metrics = append(metrics, entity.Metric{
			MType: "gauge",
			ID:    name,
			Value: &value,
		})
	}
	mu.RUnlock()

	counter := pollCount.Load()
	metrics = append(metrics, entity.Metric{
		MType: "counter",
		ID:    "PollCount",
		Delta: &counter,
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

		logger.Log.Debug("signKey is ", string(signKey))
		logger.Log.Debug("hex encoded signature is ", hex.EncodeToString(sign))
	}

	buf := bytes.Buffer{}
	gw := gzip.NewWriter(&buf)
	gw.Write(data)
	err = gw.Close()
	if err != nil {
		log.Fatal(err)
	}

	body := buf.Bytes()

	if publicKeyPath != "" {
		var err error
		body, err = encrypt(body, publicKeyPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("HashSHA256", hex.EncodeToString(sign)).
		SetHeader("X-Real-IP", "127.0.0.1").
		SetBody(body).
		Post("/updates/")
	if err != nil {
		logger.Log.Error(err)
		return
	}

	// logger.Log.Debug(resp.Header().Get("Content-Encoding"))
	logger.Log.Debugf("response code %d", resp.StatusCode())
	logger.Log.Debugf("response body %d", string(resp.Body()))
}

func CollectMemMetrics(gauge map[string]float64, mu *sync.RWMutex) {
	v, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Error(err)
		return
	}

	cpu, err := cpu.Percent(0, false)
	if err != nil {
		logger.Log.Error(err)
		return
	}

	var allCPUutil float64
	for i, val := range cpu {
		log.Println("cpu ", i, " ", val)
		allCPUutil += val
	}
	log.Println("All CPUS ", allCPUutil)

	mu.Lock()

	gauge["TotalMemory"] = float64(v.Total)
	gauge["FreeMemory"] = float64(v.Free)
	gauge["CPUutilization1"] = cpu[0]

	mu.Unlock()

	log.Println("MEM METRICS COLLECTED")
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

func SendMetricsJSON(gaugeMap map[string]float64, pollCount *int64, req *resty.Request) {
	for name, val := range gaugeMap {
		var result entity.Metric
		_, err := req.
			SetBody(&entity.Metric{
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
	var result entity.Metric
	_, err := req.
		SetBody(&entity.Metric{
			ID:    "PollCount",
			MType: "counter",
			Delta: pollCount,
		}).
		SetResult(&result).
		Post("/update/")

	if err != nil {
		return
	}
	log.Printf("\nUpdated to %v\n", result)
}

func SendMetrics(gaugeMap map[string]float64, pollCount *int64, client *resty.Client) {
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

	requestURL := fmt.Sprintf("%s/update/counter/PollCount/%d", client.BaseURL, *pollCount)
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

func CollectMetrics(gaugeMap map[string]float64, mu *sync.RWMutex) {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	mu.Lock()

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

	mu.Unlock()
	log.Printf("METRICS COLLECTED \n\n")
}

func SendTestGetJSON(req *resty.Request) {
	var result entity.Metric
	_, err := req.
		SetHeader("Accept-Encoding", "gzip").
		SetBody(&entity.Metric{
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

func encrypt(plainText []byte, keyPath string) ([]byte, error) {

	logger.Log.Debugf("palin text lenght is %d", len(plainText))

	publicKeyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		logger.Log.Error(err)
		return nil, err
	}

	logger.Log.Debugf("public key lenght is %d", len(publicKeyPEM))

	publicKeyBlock, _ := pem.Decode(publicKeyPEM)
	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
	if err != nil {
		logger.Log.Error(err)
		return nil, err
	}

	ciphertext, err := rsa.EncryptPKCS1v15(cryptoRand.Reader, publicKey, plainText)
	if err != nil {
		logger.Log.Error(err)
		return nil, err
	}

	return ciphertext, nil
}
