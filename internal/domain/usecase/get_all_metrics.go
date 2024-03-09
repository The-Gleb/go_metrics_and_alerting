package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	MetricSlices, err := uc.metricService.GetAllMetrics(ctx)
	if err != nil {
		return []byte{}, fmt.Errorf("GetAllMetricsJSON: %w", err)
	}

	b := new(bytes.Buffer)

	jsonMaps, err := json.Marshal(&MetricSlices)
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

	MetricSlices, err := uc.metricService.GetAllMetrics(ctx)
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

	for _, metric := range MetricSlices.Gauge {
		fmt.Fprintf(b, "<li>%s = %f</li>", metric.ID, *metric.Value)
	}
	for _, metric := range MetricSlices.Counter {
		fmt.Fprintf(b, "<li>%s = %d</li>", metric.ID, *metric.Delta)
	}

	fmt.Fprintf(b, "</ul></body></body>")

	return b.Bytes(), nil

}
