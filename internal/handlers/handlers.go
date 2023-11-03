package handlers

import (
	// "errors"
	"io"

	"log"
	"net/http"
	// "sync"
	// "github.com/The-Gleb/go_metrics_and_alerting/internal/models"

	"github.com/go-chi/chi/v5"
)

// type Repositiries interface {
// 	UpdateMetric(mType, mName, mValue string) error
// 	GetMetric(mType, mName string) (string, error)
// 	GetAllMetrics() (*sync.Map, *sync.Map)
// 	UpdateGauge(name string, value float64)
// 	UpdateCounter(name string, value int64)
// }

type App interface {
	UpdateMetricFromJSON(body io.Reader) ([]byte, error)
	UpdateMetricFromParams(mType, mName, mValue string) ([]byte, error)
	GetMetricFromParams(mType, mName string) ([]byte, error)
	GetMetricFromJSON(body io.Reader) ([]byte, error)
	GetAllMetricsHTML() []byte
	GetAllMetricsJSON() ([]byte, error)
	PingDB() error
	// ParamsToStruct(mType, mName, mValue string) (models.Metrics, error)
}

type handlers struct {
	app App
}

func New(app App) *handlers {
	return &handlers{
		app: app,
	}
}

func (handlers *handlers) PingDB(rw http.ResponseWriter, r *http.Request) {
	err := handlers.app.PingDB()
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (handlers *handlers) UpdateMetric(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	mType := chi.URLParam(r, "mType")
	mName := chi.URLParam(r, "mName")
	mValue := chi.URLParam(r, "mValue")
	body, err := handlers.app.UpdateMetricFromParams(mType, mName, mValue)
	// log.Printf("Body is:\n%s", body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	rw.Write(body)
}

func (handlers *handlers) UpdateMetricJSON(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	body, err := handlers.app.UpdateMetricFromJSON(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	rw.Write(body)
}

func (handlers *handlers) GetMetric(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	mType := chi.URLParam(r, "mType")
	mName := chi.URLParam(r, "mName")
	resp, err := handlers.app.GetMetricFromParams(mType, mName)
	// log.Printf("Body is: \n%s\n", resp)
	// log.Printf("Error is: \n%v\n", err)

	if err != nil {
		if err.Error() == "metric was not found" {
			rw.WriteHeader(http.StatusNotFound)
			http.Error(rw, err.Error(), http.StatusNotFound)
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
	}

	rw.Write(resp)
}

func (handlers *handlers) GetMetricJSON(rw http.ResponseWriter, r *http.Request) {

	rw.Header().Set("Content-Type", "application/json")

	resp, err := handlers.app.GetMetricFromJSON(r.Body)
	log.Printf("ОШИБКААа %v", err)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, err.Error(), http.StatusNotFound)
	}
	// rw.WriteHeader(http.StatusOK)
	rw.Write(resp)
}

func (handlers *handlers) GetAllMetricsJSON(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	body, err := handlers.app.GetAllMetricsJSON()
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		http.Error(rw, err.Error(), http.StatusNotFound)
	}
	rw.Write(body)
}

func (handlers *handlers) GetAllMetricsHTML(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html")

	body := handlers.app.GetAllMetricsHTML()
	// log.Println(body)

	// rw.WriteHeader(http.StatusOK)
	rw.Write(body)

}
