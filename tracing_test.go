package paramstore

import (
	"context"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func Test_spanErrors(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)))
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	for name, ssm := range map[string]string{"success": "success", "error": "error"} {
		recorder.Reset()
		c := &Client{
			tracerName: "paramstore",
			logger:     slog.Default(),
			batchSize:  1,
			ssmsvc:     &mockSSMClient{GetParametersFunc: mockGetParameters(ssm)},
		}
		_, _ = c.GetMultiple(context.Background(), validTestdata.toParameter().Name)

		spans := recorder.Ended()
		if len(spans) != 1 {
			t.Fatalf("%s: got %d spans, want 1", name, len(spans))
		}
		got := spans[0].Status().Code
		want := codes.Unset
		if name == "error" {
			want = codes.Error
		}
		if got != want {
			t.Errorf("%s: span status = %v, want %v", name, got, want)
		}
	}
}
