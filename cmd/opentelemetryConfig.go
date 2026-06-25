package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

type OpenTelemetryConfig struct {
	serviceName    string
	serviceVersion string
	ctx            context.Context
}

func initTracer(config OpenTelemetryConfig) (func(), error) {
	var shutdownFuncs []func(context.Context) error

	res, err := createResource(config.serviceName, config.serviceVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource %w", err)
	}

	// set from environment variables https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc#pkg-overview
	traceExporter, err := otlptracegrpc.New(config.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter %w", err)
	}
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	shutdownFuncs = append(shutdownFuncs, traceProvider.Shutdown, traceExporter.Shutdown, bsp.Shutdown)

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(newPropagator())

	metricExporter, err := otlpmetricgrpc.New(config.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter %w", err)
	}
	metricProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
	)
	shutdownFuncs = append(shutdownFuncs, metricProvider.Shutdown, metricExporter.Shutdown)
	otel.SetMeterProvider(metricProvider)

	return func() {
		for i := range shutdownFuncs {
			_ = shutdownFuncs[i](config.ctx)
		}
	}, err
}

func createResource(serviceName string, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
