package handlers

import (
	// "fmt"
	"bytes"
	"fmt"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/storage"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
)

type handlers struct {
	storage storage.Repositiries
}

type Handlers interface {
	UpdateMetric(rw http.ResponseWriter, r *http.Request)
	GetAllMetrics(rw http.ResponseWriter, r *http.Request)
	GetMetric(rw http.ResponseWriter, r *http.Request)
}

func New(store storage.Repositiries) *handlers {

	return &handlers{
		storage: store,
	}
}

func (handlers *handlers) UpdateMetric(rw http.ResponseWriter, r *http.Request) {
	// CHECK IT
	rw.Header().Set("Content-Type", "text/plain; charset=UTF-8")

	mType := chi.URLParam(r, "mType")
	mName := chi.URLParam(r, "mName")

	switch mType {
	case "gauge":
		mValue, err := strconv.ParseFloat(chi.URLParam(r, "mValue"), 64)
		if err != nil {
			http.Error(rw, "incorrect metric value\ncannot parse to float64", http.StatusBadRequest)
			return
		}
		handlers.storage.UpdateGauge(mName, mValue)
	case "counter":
		mValue, err := strconv.ParseInt(chi.URLParam(r, "mValue"), 10, 32)
		if err != nil {
			http.Error(rw, "incorrect metric value\ncannot parse to int32", http.StatusBadRequest)
			return
		}
		handlers.storage.UpdateCounter(mName, mValue)
	default:
		http.Error(rw, "Invalid mertic type", http.StatusBadRequest)
	}
	// fmt.Println("request was processed successfuly")

}

func (h *handlers) GetMetric(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain; charset=UTF-8")

	mType := chi.URLParam(r, "mType")
	mName := chi.URLParam(r, "mName")
	var stringifiedValue string
	switch mType {
	case "gauge":
		mValue, err := h.storage.GetGauge(mName)
		if err != nil {
			http.Error(rw, "metric doesn`t exist", http.StatusNotFound)
		}
		stringifiedValue = fmt.Sprintf("%v", mValue)
	case "counter":
		mValue, err := h.storage.GetCounter(mName)
		if err != nil {
			http.Error(rw, "metric doesn`t exist", http.StatusNotFound)
		}
		stringifiedValue = fmt.Sprintf("%v", mValue)
	}
	io.WriteString(rw, stringifiedValue)
}

func (h *handlers) GetAllMetrics(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")

	gaugeMap, counterMap := h.storage.GetAllMetrics()

	b := new(bytes.Buffer)
	fmt.Fprintf(b, `
	<html>
		<head>
			<meta charset="utf-8">
			<title>list-style-type</title>
			<style>
				ul {
					list-style-type: none;
				}
			</style>
		</head>
		<body>
		<ul>`)
	for key, value := range gaugeMap {
		fmt.Fprintf(b, "<li>%s = %f</li>", key, value)
	}
	for key, value := range counterMap {
		fmt.Fprintf(b, "<li>%s = %d</li>", key, value)
	}
	fmt.Fprintf(b, "</ul></body></body>")
	io.WriteString(rw, b.String())
}
