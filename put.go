package paramstore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	multierror "github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/otel"
)

// Put uploads one or more params to paramstore.
func (c *Client) Put(ctx context.Context, parameters Parameters) (errs error) {

	// setup tracing.
	newCtx, span := otel.Tracer(c.tracerName).Start(ctx, "Put")
	defer span.End()

	for _, p := range parameters {

		// setup input.
		in := &ssm.PutParameterInput{
			Name:      aws.String(p.Name),
			Value:     aws.String(p.Value),
			Type:      types.ParameterType(p.Type),
			Overwrite: aws.Bool(p.Overwrite),
		}

		// add key id, if available.
		if c.keyId != "" {
			in.KeyId = aws.String(c.keyId)
		}

		// put parameter.
		_, err := c.ssmsvc.PutParameter(newCtx, in)
		if err != nil {
			c.logger.Error(
				"failed to put parameter",
				"error", err,
				"name", *in.Name,
				"type", string(in.Type),
				"overwrite", *in.Overwrite,
			)
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}
