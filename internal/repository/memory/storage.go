package memory

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository"
)

var (
	ErrInvalidMetricValueFloat64 error = errors.New("incorrect metric value\ncannot parse to float64")
	ErrInvalidMetricValueInt64   error = errors.New("incorrect metric value\ncannot parse to int64")
	ErrInvalidMetricType         error = errors.New("invalid mertic type")
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

func (s *storage) UpdateGauge(ctx context.Context, metric entity.Metric) (entity.Metric, error) {
	s.gauge.Store(metric.ID, metric.Value)

	return metric, nil
}

func (s *storage) UpdateCounter(ctx context.Context, metric entity.Metric) (entity.Metric, error) {
	val, _ := s.counter.LoadOrStore(metric.ID, new(atomic.Int64))
	val.(*atomic.Int64).Add(*metric.Delta)

	valPtr := val.(*atomic.Int64).Load()
	metric.Delta = &valPtr
	return metric, nil
}

func (s *storage) GetGauge(ctx context.Context, dto entity.GetMetricDTO) (entity.Metric, error) {
	val, ok := s.gauge.Load(dto.ID)
	if !ok {
		return entity.Metric{}, repository.ErrNotFound
	}
	floatVal, ok := val.(*float64)
	if !ok {
		return entity.Metric{}, fmt.Errorf("error to covert value from map to *float64")
	}

	return entity.Metric{
		MType: "gauge",
		ID:    dto.ID,
		Value: floatVal,
	}, nil
}

func (s *storage) GetCounter(ctx context.Context, dto entity.GetMetricDTO) (entity.Metric, error) {
	val, ok := s.counter.Load(dto.ID)
	if !ok {
		return entity.Metric{}, repository.ErrNotFound
	}
	v := val.(*atomic.Int64).Load()

	return entity.Metric{
		MType: dto.MType,
		ID:    dto.ID,
		Delta: &v,
	}, nil
}

func (s *storage) UpdateMetricSet(ctx context.Context, metrics []entity.Metric) (int64, error) {
	var updated int64
	newGauge := *CopySyncMap(&s.gauge)
	newCounter := *CopySyncMap(&s.counter)

	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				return 0, repository.ErrInvalidMetricStruct
			}

			newGauge.Store(metric.ID, metric.Value)

			updated++
		case "counter":
			if metric.Delta == nil {
				return 0, repository.ErrInvalidMetricStruct
			}

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

func (s *storage) GetAllMetrics(ctx context.Context) (entity.MetricSlices, error) {
	newGauge := make([]entity.Metric, 0)
	s.gauge.Range(func(key any, value any) bool {
		metric := entity.Metric{
			MType: "gauge",
			ID:    key.(string),
			Value: value.(*float64),
		}

		newGauge = append(newGauge, metric)

		return true
	})

	newCounter := make([]entity.Metric, 0)
	s.counter.Range(func(key any, value any) bool {
		v := value.(*atomic.Int64).Load()
		metric := entity.Metric{
			MType: "counter",
			ID:    key.(string),
			Delta: &v,
		}

		newCounter = append(newCounter, metric)

		return true
	})

	return entity.MetricSlices{
		Gauge:   newGauge,
		Counter: newCounter,
	}, nil
}

func (s *storage) GetMetric(mType, mName string) (string, error) {
	switch mType {
	case "gauge":
		if val, ok := s.gauge.Load(mName); ok {
			if fval, ok := val.(*float64); ok {
				return fmt.Sprintf("%v", *fval), nil
			}
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
	return "", repository.ErrNotFound
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

		val.(*atomic.Int64).Add(mValue)

	default:
		return ErrInvalidMetricType
	}
	return nil
}
