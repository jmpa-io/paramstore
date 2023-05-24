package main

import (
	"context"
	"os"

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
	newCtx, span := otel.Tracer(h.name).Start(ctx, "run")
	defer span.End()

	params := []paramstore.Param{
		{
			Name:  "/jordan-test",
			Value: "hello",
			Type:  paramstore.ParamTypeSecureString,
		},
	}
	if errs := h.paramstoresvc.Put(newCtx, params...); errs != 0 {
		os.Exit(1)
	}

	// pull param.
	h.paramstoresvc.Get()
}
