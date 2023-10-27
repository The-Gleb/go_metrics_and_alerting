package storage

import (
	// "errors"
	"fmt"
	// "net/http"
	"strconv"
	"sync"
	"sync/atomic"
)

type StorageError struct {
	ErrorString string
}

func (err *StorageError) Error() string { return err.ErrorString }

var (
	ErrInvalidMetricValueFloat64 = &StorageError{"incorrect metric value\ncannot parse to float64"}
	ErrInvalidMetricValueInt64   = &StorageError{"incorrect metric value\ncannot parse to float64"}
	ErrInvalidMetricType         = &StorageError{"invalid mertic type"}
	ErrMetricDoesntExist         = &StorageError{"metric doesn`t exist"}
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
	return nil, ErrMetricDoesntExist
}

func (s *storage) GetCounter(name string) (*int64, error) {
	val, ok := s.counter.Load(name)
	if ok {
		v := val.(*atomic.Int64).Load()
		return &v, nil
	}
	return nil, ErrMetricDoesntExist
}

func (s *storage) GetAllMetrics() (*sync.Map, *sync.Map) {
	var newGauge sync.Map
	s.gauge.Range(func(key any, value any) bool {
		newGauge.Store(key, value)
		return true
	})
	var newCounter sync.Map
	s.counter.Range(func(key any, value any) bool {
		newCounter.Store(key, value)
		return true
	})
	return &newGauge, &newCounter
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
	return "", ErrMetricDoesntExist
}
