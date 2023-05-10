package paramstore

import (
	"context"

	"go.opentelemetry.io/otel"
)

type Client struct {
}

func New(ctx context.Context) (*Client, error) {

	// setup tracing.
	_, span := otel.Tracer("paramstore").Start(ctx, "New")
	defer span.End()

	return &Client{}, nil
}

func (c *Client) Get() {
	// TODO
}
