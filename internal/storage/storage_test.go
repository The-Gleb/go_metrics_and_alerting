package storage

import (
// "errors"
// "reflect"
// "sync"
// "testing"

// "github.com/stretchr/testify/assert"
)

// Do not cover edge cases(empty name or value) as above layer won`t accept it`

// func Test_storage_UpdateGauge(t *testing.T) {
// 	testStorage := New()
// 	type args struct {
// 		name  string
// 		value float64
// 	}

// 	tests := []struct {
// 		name string
// 		s    *storage
// 		args args
// 	}{
// 		{
// 			name: "pos test #1",
// 			s:    testStorage,
// 			args: args{"alloc", 23.23},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.s.UpdateGauge(tt.args.name, tt.args.value)
// 			assert.Contains(t, tt.s.gauge, tt.args.name, tt.args.value)
// 		})
// 	}
// }

// func Test_storage_UpdateCounter(t *testing.T) {
// 	testStorage := New()
// 	type args struct {
// 		name  string
// 		value int64
// 	}
// 	// type want struct {

// 	// }
// 	tests := []struct {
// 		name string
// 		s    *storage
// 		args args
// 	}{
// 		{
// 			name: "pos test #1",
// 			s:    testStorage,
// 			args: args{"alloc", 22342424},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.s.UpdateCounter(tt.args.name, tt.args.value)
// 			assert.Contains(t, tt.s.counter, tt.args.name, tt.args.value)
// 		})
// 	}
// }

// func Test_storage_GetGauge(t *testing.T) {
// 	s := New()
// 	s.gauge["Alloc"] = 234.0

// 	tests := []struct {
// 		name  string
// 		s     *storage
// 		mName string
// 		want  float64
// 		err   error
// 	}{
// 		{
// 			name:  "pos test #1",
// 			s:     s,
// 			mName: "Alloc",
// 			want:  234.0,
// 			err:   nil,
// 		},
// 		{
// 			name:  "neg test #2",
// 			s:     s,
// 			mName: "metric",
// 			want:  0.0,
// 			err:   errors.New("metric doesn`t exist"),
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			val, err := s.GetGauge(tt.mName)
// 			if err != nil {
// 				assert.Equal(t, tt.err, err)
// 			}
// 			assert.Equal(t, tt.want, val)
// 		})
// 	}
// }

// func Test_storage_GetCounter(t *testing.T) {
// 	s := New()
// 	s.counter["PollCounter"] = 234

// 	tests := []struct {
// 		name  string
// 		s     *storage
// 		mName string
// 		want  int64
// 		err   error
// 	}{
// 		{
// 			name:  "pos test #1",
// 			s:     s,
// 			mName: "PollCounter",
// 			want:  234,
// 			err:   nil,
// 		},
// 		{
// 			name:  "neg test #2",
// 			s:     s,
// 			mName: "metric",
// 			want:  0,
// 			err:   errors.New("metric doesn`t exist"),
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			val, err := s.GetCounter(tt.mName)
// 			if err != nil {
// 				assert.Equal(t, tt.err, err)
// 			}
// 			assert.Equal(t, tt.want, val)
// 		})
// 	}
// }

// func Test_storage_GetAllMetrics(t *testing.T) {
// 	tests := []struct {
// 		name  string
// 		s     *storage
// 		want  map[string]float64
// 		want1 map[string]int64
// 	}{
// 		{
// 			name: "pos test #1",
// 			s: &storage{
// 				gauge:   map[string]float64{"Alloc": 123.4, "Malloc": 567.8},
// 				counter: map[string]int64{"PollCounter": 10},
// 			},
// 			want:  map[string]float64{"Alloc": 123.4, "Malloc": 567.8},
// 			want1: map[string]int64{"PollCounter": 10},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, got1 := tt.s.GetAllMetrics()
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("storage.GetAllMetrics() got = %v, want %v", got, tt.want)
// 			}
// 			if !reflect.DeepEqual(got1, tt.want1) {
// 				t.Errorf("storage.GetAllMetrics() got1 = %v, want %v", got1, tt.want1)
// 			}
// 		})
// 	}
// }

//sync.Map
// func Test_storage_GetAllMetrics(t *testing.T) {
// 	gm := sync.Map{}
// 	gm.Store("Alloc", 123.4)
// 	gm.Store("Malloc", 567.8)
// 	tests := []struct {
// 		name  string
// 		s     *storage
// 		want  map[string]float64
// 		want1 map[string]int64
// 	}{
// 		{
// 			name: "pos test #1",
// 			s: &storage{
// 				gauge:   gm,
// 				counter: sync.Map{},//{"PollCounter": 10},
// 			},
// 			want:  map[string]float64{"Alloc": 123.4, "Malloc": 567.8},
// 			want1: map[string]int64{"PollCounter": 10},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, got1 := tt.s.GetAllMetrics()
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("storage.GetAllMetrics() got = %v, want %v", got, tt.want)
// 			}
// 			if !reflect.DeepEqual(got1, tt.want1) {
// 				t.Errorf("storage.GetAllMetrics() got1 = %v, want %v", got1, tt.want1)
// 			}
// 		})
// 	}
// }
