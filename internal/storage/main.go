package storage

import (
	"errors"
)

type storage struct {
	gauge   map[string]float64
	counter map[string]int64
}

type Repositiries interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetAllMetrics() (map[string]float64, map[string]int64)
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
}

func (s *storage) UpdateGauge(name string, value float64) {
	s.gauge[name] = value
}

func (s *storage) UpdateCounter(name string, value int64) {
	s.counter[name] += value
}

// TODO: fix GetAll()
func (s *storage) GetAllMetrics() (map[string]float64, map[string]int64) {
	return s.gauge, s.counter
}

func (s *storage) GetGauge(name string) (float64, error) {
	val, ok := s.gauge[name]
	if ok {
		return val, nil
	}
	return 0, errors.New("metric doesn`t exist")
}

func (s *storage) GetCounter(name string) (int64, error) {
	val, ok := s.counter[name]
	if ok {
		return val, nil
	}
	return 0, errors.New("metric doesn`t exist")
}

func New() *storage {

	return &storage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}
