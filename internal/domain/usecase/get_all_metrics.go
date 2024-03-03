package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/pkg/utils/retry"
)

type getAllMetricsUsecase struct {
	metricService MetricService
}

func NewGetAllMetricsUsecase(ms MetricService) *getAllMetricsUsecase {
	return &getAllMetricsUsecase{
		metricService: ms,
	}
}

func (uc *getAllMetricsUsecase) GetAllMetricsJSON(ctx context.Context) ([]byte, error) {

	var metricMaps entity.MetricsMaps
	var err error
	err = retry.DefaultRetry(context.TODO(), func(ctx context.Context) error {
		metricMaps, err = uc.metricService.GetAllMetrics(ctx)
		return err
	})
	if err != nil {
		return []byte{}, fmt.Errorf("GetAllMetricsJSON: %w", err)
	}

	b := new(bytes.Buffer)

	jsonMaps, err := json.Marshal(&metricMaps)
	if err != nil {
		return []byte{}, fmt.Errorf("GetAllMetricsJSON: %w", err)
	}

	_, err = b.Write(jsonMaps)
	if err != nil {
		return []byte{}, fmt.Errorf("GetAllMetricsJSON: %w", err)
	}

	return b.Bytes(), nil

}

func (uc *getAllMetricsUsecase) GetAllMetricsHTML(ctx context.Context) ([]byte, error) {

	var metricMaps entity.MetricsMaps
	var err error
	err = retry.DefaultRetry(ctx, func(ctx context.Context) error {
		metricMaps, err = uc.metricService.GetAllMetrics(ctx)
		return err
	})
	if err != nil {
		return []byte{}, fmt.Errorf("GetAllMetricsHTML: %w", err)
	}

	b := new(bytes.Buffer)
	fmt.Fprintf(b, `
		<html>
			<head>
				<meta charset="utf-8">
				<title>list-style-type</title>
				<style>
					ul {
						list-style-type: none;
					}
				</style>
			</head>
			<body>
			<ul>`)

	for _, metric := range metricMaps.Gauge {
		fmt.Fprintf(b, "<li>%s = %f</li>", metric.ID, *metric.Value)
	}
	for _, metric := range metricMaps.Counter {
		fmt.Fprintf(b, "<li>%s = %d</li>", metric.ID, *metric.Delta)
	}

	fmt.Fprintf(b, "</ul></body></body>")

	return b.Bytes(), nil

}
