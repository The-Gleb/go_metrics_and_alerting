package memory

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/models"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repositories"
)

var (
	ErrInvalidMetricValueFloat64 error = errors.New("incorrect metric value\ncannot parse to float64")
	ErrInvalidMetricValueInt64   error = errors.New("incorrect metric value\ncannot parse to int64")
	ErrInvalidMetricType         error = errors.New("invalid mertic type")
	// ErrMetricNotFound            error = errors.New(("metric was not found"))
)

type storage struct {
	gauge   sync.Map
	counter sync.Map
}

func New() *storage {
	return &storage{
		gauge:   sync.Map{},
		counter: sync.Map{},
	}
}

func (s *storage) UpdateMetric(mType, mName, mValue string) error {
	switch mType {
	case "gauge":
		mValue, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return ErrInvalidMetricValueFloat64
		}
		s.gauge.Store(mName, &mValue)
	case "counter":
		mValue, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return ErrInvalidMetricValueInt64
		}

		val, _ := s.counter.LoadOrStore(mName, new(atomic.Int64))
		// atomic.AddInt64(val.(*int64), mValue)
		val.(*atomic.Int64).Add(mValue)

	default:
		return ErrInvalidMetricType
	}
	return nil
}

func (s *storage) UpdateGauge(ctx context.Context, metricObj models.Metrics) error {
	s.gauge.Store(metricObj.ID, metricObj.Value)
	return nil
}

func (s *storage) UpdateCounter(ctx context.Context, metricObj models.Metrics) error {
	val, _ := s.counter.LoadOrStore(metricObj.ID, new(atomic.Int64))
	val.(*atomic.Int64).Add(*metricObj.Delta)
	return nil
}

// TODO: check
func (s *storage) GetGauge(ctx context.Context, metricObj models.Metrics) (models.Metrics, error) {
	val, ok := s.gauge.Load(metricObj.ID)
	if ok {
		metricObj.Value = val.(*float64)
		return metricObj, nil
	}
	return metricObj, repositories.ErrNotFound
}

func (s *storage) GetCounter(ctx context.Context, metricObj models.Metrics) (models.Metrics, error) {
	val, ok := s.counter.Load(metricObj.ID)
	if ok {
		v := val.(*atomic.Int64).Load()
		metricObj.Delta = &v
		return metricObj, nil
	}
	return metricObj, repositories.ErrNotFound
}

func (s *storage) UpdateMetricSet(ctx context.Context, metrics []models.Metrics) (int64, error) {
	var updated int64
	newGauge := *CopySyncMap(&s.gauge)
	newCounter := *CopySyncMap(&s.counter)
	// var newCounter sync.Map
	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			newGauge.Store(metric.ID, metric.Value)

			updated++
		case "counter":
			val, _ := newCounter.LoadOrStore(metric.ID, new(atomic.Int64))
			val.(*atomic.Int64).Add(*metric.Delta)
			updated++
		default:
			return 0, fmt.Errorf("invalid mertic type: %s", metric.MType)
		}
	}
	s.gauge = *CopySyncMap(&newGauge)
	s.counter = *CopySyncMap(&newCounter)
	return updated, nil
}

func (s *storage) GetAllMetrics(ctx context.Context) ([]models.Metrics, []models.Metrics, error) {
	newGauge := make([]models.Metrics, 0)
	s.gauge.Range(func(key any, value any) bool {
		metric := models.Metrics{
			MType: "gauge",
			ID:    key.(string),
			Value: value.(*float64),
		}
		newGauge = append(newGauge, metric)
		return true
	})
	newCounter := make([]models.Metrics, 0)
	s.counter.Range(func(key any, value any) bool {
		v := value.(*atomic.Int64).Load()
		metric := models.Metrics{
			MType: "counter",
			ID:    key.(string),
			Delta: &v,
		}
		newCounter = append(newCounter, metric)
		return true
	})
	return newGauge, newCounter, nil
}

func (s *storage) GetMetric(mType, mName string) (string, error) {
	switch mType {
	case "gauge":
		val, ok := s.gauge.Load(mName)
		if ok {
			return fmt.Sprintf("%v", val), nil
		}
	case "counter":
		val, ok := s.counter.Load(mName)
		if ok {
			v := val.(*atomic.Int64).Load()
			return fmt.Sprintf("%d", v), nil
		}
	default:
		return "", ErrInvalidMetricType
	}
	return "", repositories.ErrNotFound
}

func (s *storage) PingDB() error {
	return nil
}

func CopySyncMap(m *sync.Map) *sync.Map {
	var cp sync.Map

	m.Range(func(k, v any) bool {
		cp.Store(k, v)

		return true
	})

	return &cp
}
