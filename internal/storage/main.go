package storage

import ()

type storage struct {
	gauge   map[string]float64
	counter map[string]int64
}

type Repositiries interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}

func (s *storage) UpdateGauge(name string, value float64) {
	s.gauge[name] = value
}

func (s *storage) UpdateCounter(name string, value int64) {
	s.counter[name] = value
}

func New() *storage {

	return &storage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}
