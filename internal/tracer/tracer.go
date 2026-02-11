package tracer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Conf struct {
	Enable bool   `yaml:"enable"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
}

type InfoService struct {
	Name    string
	Version string
}

func Init(ctx context.Context, conf *Conf, serviceName string) (*trace.TracerProvider, error) {
	// Create and configure the Trace exporter
	exporterEndpoint := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(exporterEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("TraceInit: %w", err)
	}

	// labels/tags/resources that are common to all traces.
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
	)

	// Create and configure the TracerProvider exporter using the
	// newly-created exporters.
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
		// set the sampling rate based on the parent span to 60%
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(0.6))),
	)

	// Now we can register tp as the otel trace provider.
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, // W3C Trace Context format; https://www.w3.org/TR/trace-context/
		),
	)

	return tp, nil
}
