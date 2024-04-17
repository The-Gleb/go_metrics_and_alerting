package v1

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/service"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/usecase"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/repository/memory"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func Test_updateMetricSetHandler_ServeHTTP(t *testing.T) {
	s := memory.New()
	metricService := service.NewMetricService(s)
	updateMetricSetUsecase := usecase.NewUpdateMetricSetUsecase(metricService, nil)
	updateMetricSetHandler := NewUpdateMetricSetHandler(updateMetricSetUsecase)

	router := chi.NewRouter()
	updateMetricSetHandler.AddToRouter(router)
	ts := httptest.NewServer(router)
	defer ts.Close()

	validJSONBody := `[
		{
			"id": "HeapAlloc",
			"type": "gauge",
			"value": 3782369280
		},
		{
			"id": "PollCount",
			"type": "counter",
			"delta": 123
		}
	]`

	type want struct {
		code int
	}
	tests := []struct {
		name string
		body json.RawMessage
		want want
	}{
		{
			name: "normal",
			body: json.RawMessage(validJSONBody),
			want: want{200},
		},
		{
			name: "request with invalid body",
			body: json.RawMessage([]byte("some invalid body")),
			want: want{400},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, b := testRequest(t, ts, "POST", "/updates/", tt.body)
			defer resp.Body.Close()

			t.Log(b)

			require.Equal(t, tt.want.code, resp.StatusCode)
			if tt.want.code != 200 {
				return
			}
		})
	}
}
