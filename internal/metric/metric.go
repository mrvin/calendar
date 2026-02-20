package metric

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Conf struct {
	Enable bool   `env:"METRIC_ENABLE" yaml:"enable"`
	Host   string `env:"METRIC_HOST"   yaml:"host"`
	Port   int    `env:"METRIC_PORT"   yaml:"port"`
}

func Init(ctx context.Context, conf *Conf, serviceName string) (*sdkmetric.MeterProvider, error) {
	// Create and configure Metric exporter
	exporterEndpoint := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(exporterEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("MetricInit: %w", err)
	}

	// labels/tags/resources that are common to all metrics.
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
	)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(
			// collects and exports metric data every 30 seconds.
			sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(30*time.Second)),
		),
	)

	// Now we can register it as the otel meter provider.
	otel.SetMeterProvider(mp)

	return mp, nil
}
