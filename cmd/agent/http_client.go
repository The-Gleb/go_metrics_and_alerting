package main

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	"github.com/go-resty/resty/v2"
)

type httpClient struct {
	c             *resty.Client
	publicKeyPath string
	signKey       []byte
}

func NewHTTPClient(
	address string,
	signKey []byte,
	publicKeyPath string,
) (*httpClient, error) {
	baseURL := fmt.Sprintf("http://%s", address)
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

	return &httpClient{
		c:             client,
		publicKeyPath: publicKeyPath,
		signKey:       signKey,
	}, nil
}

func (c *httpClient) SendMetricSet(metrics *metricsMap) {
	metricStructs := make([]entity.Metric, 0)

	metrics.mu.RLock()
	for name, value := range metrics.Gauge {
		metricStructs = append(metricStructs, entity.Metric{
			MType: "gauge",
			ID:    name,
			Value: &value,
		})
	}
	metrics.mu.RUnlock()

	counter := metrics.PollCount.Load()
	metricStructs = append(metricStructs, entity.Metric{
		MType: "counter",
		ID:    "PollCount",
		Delta: &counter,
	})

	data, err := json.Marshal(metricStructs)
	if err != nil {
		log.Fatal(err)
	}

	logger.Log.Debug("sent body is", string(data))

	var sign []byte
	if len(c.signKey) > 0 {
		sign, err = hash(data, c.signKey)
		if err != nil {
			log.Fatal(err)
		}

		logger.Log.Debug("signKey is ", string(c.signKey))
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

	if c.publicKeyPath != "" {
		var err error
		body, err = encrypt(body, c.publicKeyPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	resp, err := c.c.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("HashSHA256", hex.EncodeToString(sign)).
		SetHeader("X-Real-IP", "127.0.0.1").
		SetBody(body).
		Post("/updates/")
	if err != nil {
		log.Fatal(err)
		return
	}

	// logger.Log.Debug(resp.Header().Get("Content-Encoding"))
	logger.Log.Debugf("response code %d", resp.StatusCode())
	logger.Log.Debugf("response body %d", string(resp.Body()))

}

func (c *httpClient) SendMetricsJSON(metrics *metricsMap) {
	pollCount := metrics.PollCount.Load()

	for name, val := range metrics.Gauge {
		var result entity.Metric
		_, err := c.c.R().
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
	_, err := c.c.R().
		SetBody(&entity.Metric{
			ID:    "PollCount",
			MType: "counter",
			Delta: &pollCount,
		}).
		SetResult(&result).
		Post("/update/")

	if err != nil {
		return
	}
	log.Printf("\nUpdated to %v\n", result)
}

func (c *httpClient) SendMetrics(metrics *metricsMap) {
	pollCount := metrics.PollCount.Load()

	for name, val := range metrics.Gauge {
		requestURL := fmt.Sprintf("%s/update/gauge/%s/%f", c.c.BaseURL, name, val)

		resp, err := c.c.R().
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

	requestURL := fmt.Sprintf("%s/update/counter/PollCount/%d", c.c.BaseURL, pollCount)
	resp, err := c.c.R().
		SetHeader("Content-Type", "application/json").
		Post(requestURL)
	if err != nil {
		log.Printf("client: error making http request: %s\n", err)
		return
	}

	logger.Log.Infow("METRICS SENT - : %s\nStatus: %d\n",
		"ADDRES", c.c.BaseURL,
		"Status", resp.StatusCode(),
	)
	log.Printf("client: status code: %d\n", resp.StatusCode())
	log.Println(string(resp.Body()))
}
