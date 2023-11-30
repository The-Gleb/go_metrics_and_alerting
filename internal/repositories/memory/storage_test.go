package memory

import (
	"sync/atomic"
	"testing"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_storage_GetMetric(t *testing.T) {
	var s storage
	var counter atomic.Int64
	counter.Store(123)
	s.gauge.Store("Alloc", 123.4)
	s.counter.Store("Counter", &counter)
	type args struct {
		mType string
		mName string
	}
	tests := []struct {
		name string
		s    *storage
		args args
		want string
		err  error
	}{
		{
			name: "pos gauge test #1",
			s:    &s,
			args: args{"gauge", "Alloc"},
			want: "123.4",
			err:  nil,
		},
		{
			name: "pos counter test #2",
			s:    &s,
			args: args{"counter", "Counter"},
			want: "123",
			err:  nil,
		},
		{
			name: "neg gauge test #3",
			s:    &s,
			args: args{"gauge", "Malloc"},
			want: "",
			err:  repositories.ErrNotFound,
		},
		{
			name: "neg bad request test #4",
			s:    &s,
			args: args{"gaug", "Malloc"},
			want: "",
			err:  ErrInvalidMetricType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.s.GetMetric(tt.args.mType, tt.args.mName)
			if err != nil {
				assert.Equal(t, tt.err, err)
				return
			}
			assert.Equal(t, tt.want, val)
		})
	}
}

func Test_storage_UpdateMetric(t *testing.T) {
	var s storage
	var counter atomic.Int64
	counter.Store(123)
	s.gauge.Store("Alloc", 123.4)
	s.counter.Store("Counter", &counter)
	type args struct {
		mType  string
		mName  string
		mValue string
	}
	tests := []struct {
		name string
		s    *storage
		args args
		val  string
		err  error
	}{
		{
			name: "pos counter test #1",
			s:    &s,
			args: args{"counter", "Counter", "7"},
			val:  "130",
			err:  nil,
		},
		{
			name: "pos gauge test #2",
			s:    &s,
			args: args{"gauge", "Alloc", "123.4"},
			val:  "123.4",
			err:  nil,
		},
		{
			name: "neg gauge test #3",
			s:    &s,
			args: args{"gauge", "Alloc", "123j"},
			val:  "123.4",
			err:  ErrInvalidMetricValueFloat64,
		},
		{
			name: "neg counter test #4",
			s:    &s,
			args: args{"counter", "Counter", "123j"},
			val:  "123.4",
			err:  ErrInvalidMetricValueInt64,
		},
		{
			name: "wrong metric type test #5",
			s:    &s,
			args: args{"gaug", "Alloc", "123"},
			val:  "123.4",
			err:  ErrInvalidMetricType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.UpdateMetric(tt.args.mType, tt.args.mName, tt.args.mValue)
			if err != nil {
				assert.Equal(t, tt.err, err, "errors not equal")
				return
			}

			val, _ := tt.s.GetMetric(tt.args.mType, tt.args.mName)
			assert.Equal(t, tt.val, val, "wrong value")
		})
	}
}
