package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Do not cover edge cases(empty name or value) as above layer won`t accept it`

func Test_storage_UpdateGauge(t *testing.T) {
	testStorage := New()
	type args struct {
		name  string
		value float64
	}
	// type want struct {

	// }
	tests := []struct {
		name string
		s    *storage
		args args
	}{
		{
			name: "pos test #1",
			s:    testStorage,
			args: args{"alloc", 23.23},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.UpdateGauge(tt.args.name, tt.args.value)
			assert.Contains(t, tt.s.gauge, tt.args.name, tt.args.value)
		})
	}
}

func Test_storage_UpdateCounter(t *testing.T) {
	testStorage := New()
	type args struct {
		name  string
		value int64
	}
	// type want struct {

	// }
	tests := []struct {
		name string
		s    *storage
		args args
	}{
		{
			name: "pos test #1",
			s:    testStorage,
			args: args{"alloc", 22342424},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.UpdateCounter(tt.args.name, tt.args.value)
			assert.Contains(t, tt.s.counter, tt.args.name, tt.args.value)
		})
	}
}
