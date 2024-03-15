package v1

import (
	"net/http/httptest"
	"testing"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/service"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/usecase"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/memory"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func Test_updateMetricHandler_ServeHTTP(t *testing.T) {
	s := memory.New()
	metricService := service.NewMetricService(s)
	updateMetricUsecase := usecase.NewUpdateMetricUsecase(metricService, nil)
	updateMetricHandler := NewUpdateMetricHandler(updateMetricUsecase)

	router := chi.NewRouter()
	updateMetricHandler.AddToRouter(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	type want struct {
		code int
	}
	tests := []struct {
		name    string
		address string
		mType   string
		mName   string
		mValue  string
		want    want
	}{
		{
			name:    "normal gauge test #1",
			address: "/update/gauge/Alloc/23.23",
			mType:   "gauge",
			mName:   "Alloc",
			mValue:  "23.23",
			want: want{
				code: 200,
			},
		},
		{
			name:    "first add counter test #2",
			address: "/update/counter/counter/23",
			mType:   "counter",
			mName:   "counter",
			mValue:  "23",
			want: want{
				code: 200,
			},
		},
		{
			name:    "second add counter test #3",
			address: "/update/counter/counter/7",
			mType:   "counter",
			mName:   "counter",
			mValue:  "30",
			want: want{
				code: 200,
			},
		},
		{
			name:    "name and value not sent - test #4",
			address: "/update/gauge",
			want: want{
				code: 404,
			},
		},
		{
			name:    "value not sent - test #5",
			address: "/update/gauge/nbhj",
			want: want{
				code: 404,
			},
		},
		{
			name:    "wrong metric type- test #5",
			address: "/update/gaunjh/efvefv/eefe",
			want: want{
				code: 400,
			},
		},
		{
			name:    "incorrect metric value type - test #6",
			address: "/update/gauge/alloc/string",
			want: want{
				code: 400,
			},
		},
		{
			name:    "empty metric name - test #7",
			address: "/update/gauge//123",
			want: want{
				code: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, "POST", tt.address, nil)
			defer resp.Body.Close()

			require.Equal(t, tt.want.code, resp.StatusCode)
			if tt.want.code != 200 {
				return
			}

			require.Equal(t, tt.mValue, body)

			// val, _ := h.app.GetMetricFromParams(context.Background(), tt.mType, tt.mName)
			// assert.Equal(t, tt.mValue, string(val))
		})
	}
}
