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

	// setup params.
	params := paramstore.Params{
		{
			Name:  "/paramstore-test/1",
			Value: "1",
			Type:  paramstore.SecureStringParamType,
		},
		{
			Name:  "/paramstore-test/2",
			Value: "2",
			Type:  paramstore.StringParamType,
		},
	}

	// setup logger.
	l := h.logger.With().Int("count", len(params)).Logger()

	// upload params.
	if errs := h.paramstoresvc.Put(newCtx, params); errs != 0 {
		l.Error().Int("errors", errs).Msg("failed to upload params")
		os.Exit(1)
	}
	l.Debug().Msg("successfully uploaded params")

	// download params.
	if _, errs := h.paramstoresvc.Get(newCtx, params.NamesToSliceString()...); errs != 0 {
		l.Error().Int("errors", errs).Msg("failed to download params")
		os.Exit(1)
	}
	l.Debug().Msg("successfully downloaded params")

	// delete params.
	if errs := h.paramstoresvc.Delete(newCtx, params.NamesToSliceString()...); errs != 0 {
		l.Error().Int("errors", errs).Msg("failed to delete params")
		os.Exit(1)
	}
	l.Debug().Msg("successfully deleted params")
}
