package paramstore

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	multierror "github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/otel"
)

// Delete deletes one or more params from paramstore.
func (c *Client) Delete(ctx context.Context, names ...string) (errs error) {

	// setup tracing.
	newCtx, span := otel.Tracer(c.tracerName).Start(ctx, "Delete")
	defer span.End()

	// retrieve params in batches.
	var invalid []string
	for i := 0; i < len(names); i += c.batchSize {

		// determine rolling batch size.
		size := i + c.batchSize
		if size > len(names) {
			size = len(names)
		}

		// delete params.
		in := &ssm.DeleteParametersInput{
			Names: names[i:size],
		}
		resp, err := c.ssmsvc.DeleteParameters(newCtx, in)
		if err != nil {
			c.logger.Error("failed to delete parameters",
				"error", err,
				"names", in.Names,
			)
			errs = multierror.Append(errs, err)
			continue
		}
		invalid = append(invalid, resp.InvalidParameters...)
	}

	// return errs.
	if len(invalid) > 0 {
		for _, i := range invalid {
			c.logger.Warn("found invalid parameter",
				"param", i,
			)
			errs = multierror.Append(errs, fmt.Errorf("%q is an invalid parameter", i))
		}
	}
	return errs
}
