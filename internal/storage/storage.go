package storage

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	// "reflect"
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

type Repositiries interface {
	UpdateMetric(mType, mName, mValue string) (int, error)
	GetMetric(mType, mName string) (string, int, error)
	GetAllMetrics() (*sync.Map, *sync.Map)
}

func (s *storage) UpdateMetric(mType, mName, mValue string) (int, error) {

	switch mType {
	case "gauge":
		mValue, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return http.StatusBadRequest, errors.New("incorrect metric value\ncannot parse to float64")
		}
		s.gauge.Store(mName, mValue)
	case "counter":
		mValue, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return http.StatusBadRequest, errors.New("incorrect metric value\ncannot parse to int32")
		}

		oldValue, ok := s.counter.Load(mName)
		if ok {
			newValue := oldValue.(int64) + mValue
			s.counter.Store(mName, newValue)
		} else {
			s.counter.Store(mName, mValue)
		}

	default:
		return http.StatusBadRequest, errors.New("invalid mertic type")
	}
	return http.StatusOK, nil
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

func (s *storage) GetMetric(mType, mName string) (string, int, error) {
	switch mType {
	case "gauge":
		val, ok := s.gauge.Load(mName)
		if ok {
			return fmt.Sprintf("%v", val), http.StatusOK, nil
		}
	case "counter":
		val, ok := s.counter.Load(mName)
		if ok {
			return fmt.Sprintf("%d", val), http.StatusOK, nil
		}
	default:
		return "", http.StatusBadRequest, errors.New("invalid mertic type")
	}
	return "", http.StatusNotFound, errors.New("metric doesn`t exist")
}
