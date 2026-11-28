package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

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
)

var (
	telemetryMu   sync.Mutex
	metricsServer *http.Server
	tracerProv    *trace.TracerProvider
)

// ShutdownTelemetry gracefully stops the Prometheus metrics HTTP server and flushes
// the OpenTelemetry TracerProvider.
func ShutdownTelemetry(ctx context.Context) error {
	telemetryMu.Lock()
	srv := metricsServer
	tp := tracerProv
	metricsServer = nil
	tracerProv = nil
	telemetryMu.Unlock()

	var errs error
	if srv != nil {
		if err := srv.Shutdown(ctx); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if tp != nil {
		if err := tp.Shutdown(ctx); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

type TelemetryConfig struct {
	Enabled     bool          `toml:"enabled" yaml:"enabled"`
	ServiceName string        `toml:"service_name" yaml:"service_name"`
	OTLP        OTLPConfig    `toml:"otlp" yaml:"otlp"`
	Metrics     MetricsConfig `toml:"metrics" yaml:"metrics"`
}
type MetricsConfig struct {
	Enabled bool   `toml:"enabled" yaml:"enabled"`
	Host    string `toml:"host" yaml:"host"`
}

type OTLPConfig struct {
	Enabled  bool   `toml:"enabled" yaml:"enabled"`
	Endpoint string `toml:"endpoint" yaml:"endpoint"`
}

func InitTelemetryFromConfig(config TelemetryConfig) error {
	if !config.Enabled {
		otel.SetTracerProvider(noop.NewTracerProvider())
		return nil
	}
	logger := slog.Default()
	if config.Metrics.Enabled {
		promHost := config.Metrics.Host
		lis, err := net.Listen("tcp", promHost)
		if err != nil {
			return err
		}
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		srv := &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		}
		telemetryMu.Lock()
		metricsServer = srv
		telemetryMu.Unlock()
		logger.Info("serving metrics", "host", promHost)
		go func() {
			if err2 := srv.Serve(lis); err2 != nil && !errors.Is(err2, http.ErrServerClosed) {
				logger.Error("metrics server stopped with error", "host", promHost, "error", err2)
			}
		}()
	}
	serviceName := config.ServiceName
	if config.ServiceName == "" {
		serviceName = "sqlsights"
	}
	if config.OTLP.Enabled {
		client := otlptracegrpc.NewClient(otlptracegrpc.WithEndpoint(config.OTLP.Endpoint), otlptracegrpc.WithInsecure())
		exporter, err := otlptrace.New(context.Background(), client)
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
		telemetryMu.Lock()
		tracerProv = tp
		telemetryMu.Unlock()
		otel.SetTracerProvider(tp)
		propagator := propagation.NewCompositeTextMapPropagator(
			propagation.Baggage{},
			propagation.TraceContext{},
			b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader|b3.B3SingleHeader)))
		otel.SetTextMapPropagator(propagator)
	}
	return nil
}
