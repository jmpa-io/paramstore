package main

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"

	"github.com/jmpa-io/paramstore"
)

type handler struct {

	// config.
	name        string
	version     string
	environment string

	// clients.
	paramstoresvc *paramstore.Client

	// misc.
	logger *slog.Logger
}

// run is like main but after the handler is configured.
func (h *handler) run(ctx context.Context) error {

	// setup span.
	newCtx, span := otel.Tracer(h.name).Start(ctx, "run")
	defer span.End()

	// setup params.
	params := paramstore.Parameters{
		{
			Name:  "/paramstore-test/1",
			Value: "1",
			Type:  paramstore.ParameterTypeSecureString,
		},
		{
			Name:  "/paramstore-test/2",
			Value: "2",
			Type:  paramstore.ParameterTypeString,
		},
	}

	// setup logger.
	l := h.logger.With("count", len(params))

	// upload params.
	if err := h.paramstoresvc.Put(newCtx, params); err != nil {
		l.Error("failed to upload parameters", "error", err)
		return err
	}
	l.Debug("successfully uploaded parameters")

	// download param.
	if _, err := h.paramstoresvc.Get(newCtx, params[0].Name); err != nil {
		l.Error("failed to download parameters", "error", err)
		return err
	}
	l.Debug("successfully downloaded parameter")

	// download params (multiple).
	if _, err := h.paramstoresvc.GetMultiple(newCtx, params.ToSliceString()...); err != nil {
		l.Error("failed to download parameters", "error", err)
		return err
	}
	l.Debug("successfully downloaded parameters")

	// delete params.
	if err := h.paramstoresvc.Delete(newCtx, params.ToSliceString()...); err != nil {
		l.Error("failed to delete parameters", "error", err)
		return err
	}
	l.Debug("successfully deleted parameters")
	return nil
}
