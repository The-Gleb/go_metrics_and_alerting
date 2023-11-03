package storage

import (
	"errors"
	"fmt"
	"log"

	// "net/http"
	"strconv"
	"sync"
	"sync/atomic"
)

var (
	ErrInvalidMetricValueFloat64 error = errors.New("incorrect metric value\ncannot parse to float64")
	ErrInvalidMetricValueInt64   error = errors.New("incorrect metric value\ncannot parse to int64")
	ErrInvalidMetricType         error = errors.New("invalid mertic type")
	ErrMetricNotFound            error = errors.New(("metric was not found"))
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
		s.gauge.Store(mName, mValue)
	case "counter":
		mValue, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return ErrInvalidMetricValueInt64
		}

		val, _ := s.counter.LoadOrStore(mName, new(*atomic.Int64))
		// atomic.AddInt64(val.(*int64), mValue)
		val.(*atomic.Int64).Add(mValue)

	default:
		return ErrInvalidMetricType
	}
	return nil
}

func (s *storage) UpdateGauge(name string, value float64) {
	s.gauge.Store(name, value)
}

func (s *storage) UpdateCounter(name string, value int64) {
	val, _ := s.counter.LoadOrStore(name, new(atomic.Int64))
	val.(*atomic.Int64).Add(value)
}

func (s *storage) GetGauge(name string) (*float64, error) {
	val, ok := s.gauge.Load(name)
	if ok {
		v := val.(float64)
		return &v, nil
	}
	return nil, ErrMetricNotFound
}

func (s *storage) GetCounter(name string) (*int64, error) {
	val, ok := s.counter.Load(name)
	if ok {
		v := val.(*atomic.Int64).Load()
		log.Printf("Got in storage %d ", v)
		return &v, nil
	}
	return nil, ErrMetricNotFound
}

func (s *storage) GetAllMetrics() (map[string]float64, map[string]int64) {
	newGauge := make(map[string]float64)
	s.gauge.Range(func(key any, value any) bool {
		newGauge[key.(string)] = value.(float64)
		return true
	})
	newCounter := make(map[string]int64)
	s.counter.Range(func(key any, value any) bool {
		v := value.(*atomic.Int64).Load()
		newCounter[key.(string)] = v
		return true
	})
	return newGauge, newCounter
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
	return "", ErrMetricNotFound
}

func (s *storage) PingDB() error {
	return nil
}
