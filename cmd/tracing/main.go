package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/jmpa-io/paramstore"
)

var (

	// the name of this binary.
	Name = "example"

	// the version of this binary.
	Version = "HEAD"
)

func main() {

	// setup log level.
	logLevel := os.Getenv("LOG_LEVEL")
	level := slog.LevelWarn
	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	// setup handler.
	h := &handler{

		// config.
		name:        Name,
		version:     Version,
		environment: getEnv("ENVIRONMENT", "dev"),

		// misc.
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05"))
				}
				return a
			},
		})),
	}

	// setup exporter.
	exp, err := newExporter()
	if err != nil {
		h.logger.Error("failed to setup exporter",
			"type", "grpc",
			"error", err,
		)
		os.Exit(1)
	}

	// setup trace provider.
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(newResource(h.name, h.version, h.environment)),
	)
	defer func() {
		if err := tp.Shutdown(context.TODO()); err != nil {
			h.logger.Error("failed to shutdown tracer provider", "error", err)
			os.Exit(1)
		}
	}()
	otel.SetTracerProvider(tp)

	// ---

	// setup span.
	ctx, span := otel.Tracer(h.name).Start(context.Background(), "main")
	defer span.End()

	// setup client.
	h.paramstoresvc, err = paramstore.New(ctx)
	if err != nil {
		h.logger.Error("failed to setup client", "client", "paramstore")
		os.Exit(1)
	}

	// ~start!
	h.run(ctx)
}

// newResource returns a resource describing this app.
func newResource(app, version, env string) *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(app),
			semconv.ServiceVersion(version),
			attribute.String("environment", env),
		),
	)
	return r
}

// newExporter returns a started grpc trace exporter.
//
// NOTE: set 'OTLP_EXPORTER_OTLP_TRACES_ENDPOINT' to change the endpoint from
// https://localhost:4317 to another endpoint.
func newExporter() (*otlptrace.Exporter, error) {
	return otlptracegrpc.New(
		context.TODO(),
		otlptracegrpc.WithInsecure(),
	)
}

// getEnv retrieves an environment variable value with a default fallback.
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
