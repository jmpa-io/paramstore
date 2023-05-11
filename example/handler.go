package main

import (
	"context"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"

	"github.com/myob-ops/paramstore"
)

type handler struct {

	// config.
	name, version, env string

	// clients.
	paramstoresvc *paramstore.Client

	// misc.
	logger zerolog.Logger
}

func (h *handler) run(ctx context.Context) {

	// setup span.
	_, span := otel.Tracer(h.name).Start(ctx, "run")
	defer span.End()

	// pull param.
	h.paramstoresvc.Get()
	// TODO
}
