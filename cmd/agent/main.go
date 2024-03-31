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

type metricsMap struct {
	Gauge     map[string]float64
	PollCount atomic.Int64
	mu        sync.RWMutex
}

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

	metrics := metricsMap{
		Gauge: make(map[string]float64),
	}
	metrics.PollCount.Store(1)

	var pollInterval = time.Duration(config.PollInterval * 1000000000)
	var reportInterval = time.Duration(config.ReportInterval * 1000000000)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	sendTaskCh := make(chan struct{}, 1)
	collectTaskCh := make(chan struct{}, 1)
	collectMemsTaskCh := make(chan struct{}, 1)

	client, err := NewGRPCClient(config.Address, []byte(config.SignKey), config.PublicKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < config.RateLimit; i++ {
		go func() {
			for range sendTaskCh {
				client.SendMetricSet(&metrics)
				client.GetAllMetrics()
			}
		}()
	}

	go func() {
		for range collectTaskCh {
			CollectMetrics(&metrics)
		}
	}()

	go func() {
		for range collectMemsTaskCh {
			CollectMemMetrics(&metrics)
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

func CollectMemMetrics(metrics *metricsMap) {
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

	metrics.mu.Lock()

	metrics.Gauge["TotalMemory"] = float64(v.Total)
	metrics.Gauge["FreeMemory"] = float64(v.Free)
	metrics.Gauge["CPUutilization1"] = cpu[0]

	metrics.mu.Unlock()

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

func CollectMetrics(metrics *metricsMap) {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	metrics.mu.Lock()

	metrics.Gauge["Alloc"] = float64(rtm.Alloc)
	metrics.Gauge["BuckHashSys"] = float64(rtm.BuckHashSys)
	metrics.Gauge["Frees"] = float64(rtm.Frees)
	metrics.Gauge["GCCPUFraction"] = float64(rtm.GCCPUFraction)
	metrics.Gauge["GCSys"] = float64(rtm.GCSys)
	metrics.Gauge["HeapAlloc"] = float64(rtm.HeapAlloc)
	metrics.Gauge["HeapIdle"] = float64(rtm.HeapIdle)
	metrics.Gauge["HeapInuse"] = float64(rtm.HeapInuse)
	metrics.Gauge["HeapObjects"] = float64(rtm.HeapObjects)
	metrics.Gauge["HeapReleased"] = float64(rtm.HeapReleased)
	metrics.Gauge["HeapSys"] = float64(rtm.HeapSys)
	metrics.Gauge["LastGC"] = float64(rtm.LastGC)
	metrics.Gauge["Lookups"] = float64(rtm.Lookups)
	metrics.Gauge["MCacheInuse"] = float64(rtm.MCacheInuse)
	metrics.Gauge["MCacheSys"] = float64(rtm.MCacheSys)
	metrics.Gauge["MSpanInuse"] = float64(rtm.MSpanInuse)
	metrics.Gauge["MSpanSys"] = float64(rtm.MSpanSys)
	metrics.Gauge["Mallocs"] = float64(rtm.Mallocs)
	metrics.Gauge["NextGC"] = float64(rtm.NextGC)
	metrics.Gauge["NumForcedGC"] = float64(rtm.NumForcedGC)
	metrics.Gauge["NumGC"] = float64(rtm.NumGC)
	metrics.Gauge["OtherSys"] = float64(rtm.OtherSys)
	metrics.Gauge["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	metrics.Gauge["StackInuse"] = float64(rtm.StackInuse)
	metrics.Gauge["StackSys"] = float64(rtm.StackSys)
	metrics.Gauge["Sys"] = float64(rtm.Sys)
	metrics.Gauge["TotalAlloc"] = float64(rtm.TotalAlloc)
	metrics.Gauge["RandomValue"] = rand.Float64()

	metrics.mu.Unlock()
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
