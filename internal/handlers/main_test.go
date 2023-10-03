package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_handlers_UpdateMetric(t *testing.T) {
	mockStorage := storage.New()
	mockHandlers := New(mockStorage)
	type want struct {
		code int
		// err  error
	}
	tests := []struct {
		name         string
		h            *handlers
		address      string
		mType        string
		mName        string
		gaugeValue   float64
		counterValue int64
		want         want
	}{
		{
			name:       "normal gauge test #1",
			address:    "/gauge/alloc/23.23",
			mType:      "gauge",
			mName:      "alloc",
			gaugeValue: 23.23,
			h:          mockHandlers,
			want: want{
				code: 200,
			},
		},
		{
			name:         "first add counter test #1",
			address:      "/counter/counter/23345",
			mType:        "counter",
			mName:        "counter",
			counterValue: 23345,
			h:            mockHandlers,
			want: want{
				code: 200,
			},
		},
		{
			name:         "second add counter test #1",
			address:      "/counter/counter/27",
			mType:        "counter",
			mName:        "counter",
			counterValue: 27,
			h:            mockHandlers,
			want: want{
				code: 200,
			},
		},
		{
			name:    "name and value not sent - test #3",
			address: "/gauge",
			h:       mockHandlers,
			want: want{
				code: 404,
			},
		},
		{
			name:    "value not sent - test #4",
			address: "/gauge/nbhj",
			h:       mockHandlers,
			want: want{
				code: 404,
			},
		},
		{
			name:    "wrong metric type- test #5",
			address: "/gaunjh/efvefv/eefe",
			h:       mockHandlers,
			want: want{
				code: 400,
			},
		},
		{
			name:    "nothing was sent - test #6",
			address: "",
			h:       mockHandlers,
			want: want{
				code: 404,
			},
		},
		{
			name:    "incorrect metric value type - test #6",
			address: "/gauge/alloc/string",
			h:       mockHandlers,
			want: want{
				code: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var oldCounter int64
			if tt.mType == "counter" {
				oldCounter, _ = tt.h.storage.GetCounter(tt.mName)
			}

			request := httptest.NewRequest(http.MethodPost, "/update"+tt.address, nil)
			rw := httptest.NewRecorder()
			tt.h.UpdateMetric(rw, request)

			res := rw.Result()
			defer res.Body.Close()
			// проверяем код ответа
			require.Equal(t, tt.want.code, res.StatusCode)
			if tt.want.code != 200 {
				t.Skip()
			}
			// получаем и проверяем тело запроса
			switch tt.mType {
			case "gauge":
				val, _ := tt.h.storage.GetGauge(tt.mName)
				assert.Equal(t, tt.gaugeValue, val)
			case "counter":
				val, _ := tt.h.storage.GetCounter(tt.mName)
				assert.Equal(t, tt.counterValue+oldCounter, val)
			}

		})
	}
}
