package storage

import ()

type storage struct {
	gauge   map[string]float64
	counter map[string]int64
}

type Repositiries interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetAll() (map[string]float64, map[string]int64)
	GetGauge(name string) float64
	GetCounter(name string) int64
}

func (s *storage) UpdateGauge(name string, value float64) {
	s.gauge[name] = value
}

func (s *storage) UpdateCounter(name string, value int64) {
	s.counter[name] += value
}

// TODO: fix GetAll()
func (s *storage) GetAll() (map[string]float64, map[string]int64) {
	return s.gauge, s.counter
}

func (s *storage) GetGauge(name string) float64 {
	return s.gauge[name]
}

func (s *storage) GetCounter(name string) int64 {
	return s.counter[name]
}

func New() *storage {

	return &storage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}
