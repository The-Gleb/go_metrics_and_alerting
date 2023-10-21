package handlers

import (
	// "fmt"
	"bytes"
	"fmt"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"sync"
)

type Repositiries interface {
	UpdateMetric(mType, mName, mValue string) error
	GetMetric(mType, mName string) (string, error)
	GetAllMetrics() (*sync.Map, *sync.Map)
}

type handlers struct {
	storage Repositiries
}

func New(store Repositiries) *handlers {

	return &handlers{
		storage: store,
	}
}

func (handlers *handlers) UpdateMetric(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain; charset=UTF-8")

	mType := chi.URLParam(r, "mType")
	mName := chi.URLParam(r, "mName")
	mValue := chi.URLParam(r, "mValue")
	err := handlers.storage.UpdateMetric(mType, mName, mValue)

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	rw.WriteHeader(http.StatusOK)
}

func (handlers *handlers) GetMetric(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain; charset=UTF-8")

	mType := chi.URLParam(r, "mType")
	mName := chi.URLParam(r, "mName")
	mValue, err := handlers.storage.GetMetric(mType, mName)

	if err != nil {
		switch err {
		case storage.ErrInvalidMetricType:
			http.Error(rw, err.Error(), http.StatusBadRequest)
		case storage.ErrMetricDoesntExist:
			http.Error(rw, err.Error(), http.StatusNotFound)
		}

	}
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(mValue))
}

func (handlers *handlers) GetAllMetrics(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")

	gaugeMap, counterMap := handlers.storage.GetAllMetrics()

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
	gaugeMap.Range(func(key, value any) bool {
		fmt.Fprintf(b, "<li>%s = %f</li>", key, value)
		return true
	})
	counterMap.Range(func(key, value any) bool {
		fmt.Fprintf(b, "<li>%s = %d</li>", key, value)
		return true
	})

	fmt.Fprintf(b, "</ul></body></body>")
	rw.WriteHeader(http.StatusOK)
	rw.Write(b.Bytes())
}
