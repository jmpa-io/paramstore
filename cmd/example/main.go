package main

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/jmpa-io/paramstore"
)

const (
	Name    = "example"
	Version = "head"
)

func init() {

	// logger global config.
	zerolog.TimestampFieldName = "ts"
	zerolog.MessageFieldName = "msg"
}

func main() {

	// setup log level.
	logLevel := os.Getenv("LOG_LEVEL")
	var level zerolog.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	default:
		level = zerolog.DebugLevel
	}

	// setup handler.
	h := &handler{
		config: config{
			name:    Name,
			version: Version,
			env:     getEnv("ENVIRONMENT", "dev"),
		},

		logger: zerolog.New(os.Stderr).
			With().Timestamp().Logger().Level(level),
	}

	// setup file to export traces to.
	file := getEnv("TELEMETRY_FILE", "traces.txt")
	f, err := os.Create(file)
	if err != nil {
		h.logger.Fatal().
			Err(err).
			Str("name", file).
			Msg("failed to create file")
	}
	defer f.Close()

	// setup exporter.
	exp, err := newExporter(f)
	if err != nil {
		h.logger.Fatal().
			Err(err).
			Msg("failed to setup exporter")
	}

	// setup trace provider.
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(newResource(h.name, h.env, h.version)),
	)
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			h.logger.Fatal().
				Err(err).
				Msg("failed to shutdown trace provider")
		}
	}()
	otel.SetTracerProvider(tp)

	// ---

	// setup span.
	ctx, span := otel.Tracer(h.name).Start(context.Background(), "main")
	defer span.End()

	// setup paramstore.
	h.paramstoresvc, err = paramstore.New(ctx)
	if err != nil {
		h.logger.Fatal().
			Err(err).
			Str("service", "paramstore").
			Msg("failed to setup service")
	}

	// run.
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

// newExporter returns a configured console exporter.
func newExporter(w io.Writer) (trace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		stdouttrace.WithPrettyPrint(),
	)
}

// getEnv retrieves an environment variable value with a default fallback.
func getEnv(envVar, fallback string) string {
	v := os.Getenv(envVar)
	if v == "" {
		return fallback
	}
	return v
}
