package storage

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_storage_GetMetric(t *testing.T) {
	var s storage
	s.gauge.Store("Alloc", 123.4)
	s.counter.Store("Counter", 123)
	type args struct {
		mType string
		mName string
	}
	tests := []struct {
		name  string
		s     *storage
		args  args
		want  string
		want1 int
		err   error
	}{
		{
			name:  "pos gauge test #1",
			s:     &s,
			args:  args{"gauge", "Alloc"},
			want:  "123.4",
			want1: http.StatusOK,
			err:   nil,
		},
		{
			name:  "pos counter test #2",
			s:     &s,
			args:  args{"counter", "Counter"},
			want:  "123",
			want1: http.StatusOK,
			err:   nil,
		},
		{
			name:  "neg gauge test #3",
			s:     &s,
			args:  args{"gauge", "Malloc"},
			want:  "",
			want1: http.StatusNotFound,
			err:   errors.New("metric doesn`t exist"),
		},
		{
			name:  "neg bad request test #4",
			s:     &s,
			args:  args{"gaug", "Malloc"},
			want:  "",
			want1: http.StatusBadRequest,
			err:   errors.New("invalid mertic type"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, code, err := tt.s.GetMetric(tt.args.mType, tt.args.mName)
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.want, val)
			assert.Equal(t, tt.want1, code)
		})
	}
}

func Test_storage_UpdateMetric(t *testing.T) {
	var s storage
	s.gauge.Store("Alloc", 123.4)
	s.counter.Store("Counter", 123)
	type args struct {
		mType  string
		mName  string
		mValue string
	}
	tests := []struct {
		name       string
		s          *storage
		args       args
		val        string
		statusCode int
		err        error
	}{
		{
			name:       "pos counter test #1",
			s:          &s,
			args:       args{"counter", "Counter", "7"},
			val:        "130",
			statusCode: http.StatusOK,
			err:        nil,
		},
		{
			name:       "pos gauge test #2",
			s:          &s,
			args:       args{"gauge", "Alloc", "123.4"},
			val:        "123.4",
			statusCode: http.StatusOK,
			err:        nil,
		},
		{
			name:       "neg gauge test #3",
			s:          &s,
			args:       args{"gauge", "Alloc", "123j"},
			val:        "123.4",
			statusCode: http.StatusBadRequest,
			err:        errors.New("incorrect metric value\ncannot parse to float64"),
		},
		{
			name:       "wrong metric type test #4",
			s:          &s,
			args:       args{"gaug", "Alloc", "123"},
			val:        "123.4",
			statusCode: http.StatusBadRequest,
			err:        errors.New("invalid mertic type"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, err := s.UpdateMetric(tt.args.mType, tt.args.mName, tt.args.mValue)
			assert.Equal(t, tt.statusCode, statusCode, "status codes are not equal")
			assert.Equal(t, tt.err, err, "errors not equal")
			if statusCode == 200 {
				val, _, _ := tt.s.GetMetric(tt.args.mType, tt.args.mName)
				assert.Equal(t, tt.val, val, "wrong value")
			}
		})
	}
}
