package main

import (
	"context"
	"log"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/domain/entity"
	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
	metrics "github.com/The-Gleb/go_metrics_and_alerting/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcClient struct {
	client        metrics.MetricServiceClient
	signKey       []byte
	publicKeyPath string
}

func NewGRPCClient(
	address string, signKey []byte, publicKeyPath string,
) (*grpcClient, error) {
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(),
	)
	if err != nil {
		return nil, err
	}

	c := metrics.NewMetricServiceClient(conn)

	// TODO: interceptors

	return &grpcClient{
		client:        c,
		signKey:       signKey,
		publicKeyPath: publicKeyPath,
	}, nil

}

// func (c *grpcClient) SomeInterceptor(
// 	ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn,
// 	invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
// ) error {

// }

func (c *grpcClient) SendMetricSet(metricsMap *metricsMap) {
	metricStructs := make([]*metrics.UpdateMetricRequest, 0)

	for name, value := range metricsMap.Gauge {

		m := metrics.UpdateMetricRequest{
			Type:  metrics.MetricType_GAUGE,
			Name:  name,
			Value: value,
		}

		metricStructs = append(metricStructs, &m)

	}

	metricStructs = append(metricStructs, &metrics.UpdateMetricRequest{
		Type:  metrics.MetricType_COUNTER,
		Name:  "PollCount",
		Delta: metricsMap.PollCount.Load(),
	})

	in := &metrics.UpdateMetricSetRequest{
		Metrics: metricStructs,
	}

	logger.Log.Debug("METRICS TO SEND")
	logger.Log.Debug(metricStructs)

	resp, err := c.client.UpdateMetricSet(context.Background(), in)
	if err != nil {
		log.Fatal(err)
	}

	logger.Log.Debugf("%d metrics updated", resp.GetUpdatedNum())

}

func (c *grpcClient) GetAllMetrics() ([]entity.Metric, error) {
	resp, err := c.client.GetAllMetrics(context.Background(), &metrics.GetAllMetricsRequest{})
	if err != nil {
		return nil, err
	}

	metricStructs := make([]entity.Metric, len(resp.Metrics))

	for _, m := range resp.Metrics {
		counter := m.GetCounterValue()

		metricStructs = append(metricStructs, entity.Metric{
			MType: metrics.MetricType_name[int32(m.Type)],
			ID:    m.GetName(),
			Value: &m.GaugeValue,
			Delta: &counter,
		})
	}

	return metricStructs, nil
}
