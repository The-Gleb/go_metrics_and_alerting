package memory

import (
	"context"
	"reflect"
	"testing"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository"
	"github.com/stretchr/testify/assert"
)

func Test_storage_GetMetric(t *testing.T) {
	memory := New()
	memory.UpdateMetric("gauge", "Alloc", "123.4")
	memory.UpdateMetric("counter", "Counter", "123")
	type args struct {
		mType string
		mName string
	}
	tests := []struct {
		name string
		args args
		want string
		err  error
	}{
		{
			name: "pos gauge test #1",
			args: args{"gauge", "Alloc"},
			want: "123.4",
			err:  nil,
		},
		{
			name: "pos counter test #2",
			args: args{"counter", "Counter"},
			want: "123",
			err:  nil,
		},
		{
			name: "neg gauge test #3",
			args: args{"gauge", "Malloc"},
			want: "",
			err:  repository.ErrNotFound,
		},
		{
			name: "neg bad request test #4",
			args: args{"gaug", "Malloc"},
			want: "",
			err:  ErrInvalidMetricType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := memory.GetMetric(tt.args.mType, tt.args.mName)
			if err != nil {
				assert.Equal(t, tt.err, err)
				return
			}
			assert.Equal(t, tt.want, val)
		})
	}
}

func Test_storage_UpdateMetric(t *testing.T) {
	memory := New()
	type args struct {
		mType  string
		mName  string
		mValue string
	}
	tests := []struct {
		name string
		args args
		val  string
		err  error
	}{
		{
			name: "pos new counter test #1",
			args: args{"counter", "Counter", "7"},
			val:  "7",
			err:  nil,
		},
		{
			name: "pos update counter test #2",
			args: args{"counter", "Counter", "3"},
			val:  "10",
			err:  nil,
		},
		{
			name: "pos new gauge test #3",
			args: args{"gauge", "Alloc", "123.4"},
			val:  "123.4",
			err:  nil,
		},
		{
			name: "pos update gauge test #4",
			args: args{"gauge", "Alloc", "123.4"},
			val:  "123.4",
			err:  nil,
		},
		{
			name: "neg gauge test #5",
			args: args{"gauge", "Alloc", "123j"},
			val:  "123.4",
			err:  ErrInvalidMetricValueFloat64,
		},
		{
			name: "neg counter test #6",
			args: args{"counter", "Counter", "123j"},
			val:  "123.4",
			err:  ErrInvalidMetricValueInt64,
		},
		{
			name: "wrong metric type test #7",
			args: args{"gaug", "Alloc", "123"},
			val:  "123.4",
			err:  ErrInvalidMetricType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := memory.UpdateMetric(tt.args.mType, tt.args.mName, tt.args.mValue)
			if err != nil {
				assert.Equal(t, tt.err, err, "errors not equal")
				return
			}

			val, _ := memory.GetMetric(tt.args.mType, tt.args.mName)
			assert.Equal(t, tt.val, val, "wrong value")
		})
	}
}

func Test_storage_GetAllMetrics(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		s       *storage
		args    args
		want    entity.MetricSlices
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetAllMetrics(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("storage.GetAllMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("storage.GetAllMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}
