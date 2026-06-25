package application

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type DLQMetrics interface {
	MessageStored(ctx context.Context, channel string, routingKey string)
	StoreFailed(ctx context.Context, channel string, routingKey string)
}

type otelDLQMetrics struct {
	messagesStored metric.Int64Counter
	storeFailures  metric.Int64Counter
}

func NewDLQMetrics() (DLQMetrics, error) {
	meter := otel.Meter("sms-gateway")

	messagesStored, err := meter.Int64Counter(
		"sms_gateway_dlq_messages_total",
		metric.WithDescription("Total number of messages consumed from the SMS gateway dead-letter queue and stored."),
		metric.WithUnit("{message}"),
	)
	if err != nil {
		return nil, err
	}

	storeFailures, err := meter.Int64Counter(
		"sms_gateway_dlq_store_failures_total",
		metric.WithDescription("Total number of failures while storing SMS gateway dead-letter queue messages."),
		metric.WithUnit("{failure}"),
	)
	if err != nil {
		return nil, err
	}

	return otelDLQMetrics{
		messagesStored: messagesStored,
		storeFailures:  storeFailures,
	}, nil
}

func (metrics otelDLQMetrics) MessageStored(ctx context.Context, channel string, routingKey string) {
	metrics.messagesStored.Add(ctx, 1, metric.WithAttributes(dlqMetricAttributes(channel, routingKey)...))
}

func (metrics otelDLQMetrics) StoreFailed(ctx context.Context, channel string, routingKey string) {
	metrics.storeFailures.Add(ctx, 1, metric.WithAttributes(dlqMetricAttributes(channel, routingKey)...))
}

func dlqMetricAttributes(channel string, routingKey string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("messaging.destination.name", channel),
		attribute.String("messaging.rabbitmq.routing_key", routingKey),
	}
}

type noopDLQMetrics struct{}

func (noopDLQMetrics) MessageStored(context.Context, string, string) {}
func (noopDLQMetrics) StoreFailed(context.Context, string, string)   {}
