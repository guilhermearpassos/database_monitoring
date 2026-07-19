package telemetry

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace/noop"
	"log/slog"
	"net/http"
	"time"
)

type TelemetryConfig struct {
	Enabled     bool          `toml:"enabled"`
	ServiceName string        `toml:"service_name"`
	OTLP        OTLPConfig    `toml:"otlp"`
	Metrics     MetricsConfig `toml:"metrics"`
}
type MetricsConfig struct {
	Enabled bool   `toml:"enabled"`
	Host    string `toml:"host"`
}

type OTLPConfig struct {
	Enabled  bool   `toml:"enabled"`
	Endpoint string `toml:"endpoint"`
}

func InitTelemetryFromConfig(config TelemetryConfig) error {
	if !config.Enabled {
		otel.SetTracerProvider(noop.NewTracerProvider())
		return nil
	}
	logger := slog.Default()
	if config.Metrics.Enabled {

		go func() {
			promHost := config.Metrics.Host
			mux := http.NewServeMux()
			mux.Handle("/metrics", promhttp.Handler())
			logger.Info("serving metrics on %s", promHost)
			err2 := http.ListenAndServe(promHost, mux)
			if err2 != nil {
				panic(err2)
			}
		}()
	}
	serviceName := config.ServiceName
	if config.ServiceName == "" {
		serviceName = "sqlsights"
	}
	if config.OTLP.Enabled {
		client := otlptracegrpc.NewClient(otlptracegrpc.WithEndpoint(config.OTLP.Endpoint), otlptracegrpc.WithInsecure())
		exporter, err := otlptrace.New(context.TODO(), client)
		if err != nil {
			return err
		}
		defaultResource := resource.Default()
		r, err := resource.Merge(
			resource.NewSchemaless(defaultResource.Attributes()...),
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
			))
		if err != nil {
			return err
		}
		tp := trace.NewTracerProvider(
			trace.WithBatcher(exporter, trace.WithExportTimeout(30*time.Second)),
			trace.WithResource(r))
		otel.SetTracerProvider(tp)
		propagator := propagation.NewCompositeTextMapPropagator(
			propagation.Baggage{},
			propagation.TraceContext{},
			b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader|b3.B3SingleHeader)))
		otel.SetTextMapPropagator(propagator)
	}
	return nil
}
