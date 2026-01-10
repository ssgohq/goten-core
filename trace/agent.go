package trace

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/ssgohq/goten-core/logx"
)

// StartAgent initializes OpenTelemetry tracing based on configuration.
// It returns a shutdown function that should be called when the application exits.
//
// Example:
//
//	shutdown, err := trace.StartAgent(trace.Config{
//	    Name:     "my-service",
//	    Endpoint: "http://localhost:4318",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer shutdown(context.Background())
func StartAgent(cfg Config) (func(context.Context) error, error) {
	if !cfg.IsEnabled() {
		logx.Debugw("Tracing disabled", "name", cfg.Name)
		return func(_ context.Context) error { return nil }, nil
	}

	cfg.SetDefaults()

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.Name),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create exporter
	exporter, err := createExporter(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create sampler
	var sampler sdktrace.Sampler
	if cfg.SampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.SampleRate <= 0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.SampleRate)
	}

	// Create TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(cfg.BatchTimeout),
			sdktrace.WithExportTimeout(cfg.ExportTimeout),
			sdktrace.WithMaxExportBatchSize(cfg.MaxExportBatchSize),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global TracerProvider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logx.Infow("Tracing initialized",
		"name", cfg.Name,
		"endpoint", cfg.Endpoint,
		"exporter", cfg.Exporter,
		"sampleRate", cfg.SampleRate,
	)

	return tp.Shutdown, nil
}

// createExporter creates the appropriate exporter based on configuration.
func createExporter(cfg Config) (sdktrace.SpanExporter, error) {
	switch strings.ToLower(cfg.Exporter) {
	case "stdout":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())

	case "otlp", "":
		return createOTLPExporter(cfg)

	case "jaeger":
		// Jaeger now supports OTLP protocol
		return createOTLPExporter(cfg)

	default:
		return nil, fmt.Errorf("unknown exporter type: %s", cfg.Exporter)
	}
}

// createOTLPExporter creates an OTLP HTTP exporter.
func createOTLPExporter(cfg Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(normalizeEndpoint(cfg.Endpoint)),
	}

	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.Headers))
	}

	return otlptracehttp.New(context.Background(), opts...)
}

// normalizeEndpoint removes protocol prefix from endpoint.
func normalizeEndpoint(endpoint string) string {
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "grpc://")
	return endpoint
}