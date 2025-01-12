package paramstore

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	multierror "github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/otel"
)

// Get retrieves a single param from paramstore.
func (c *Client) Get(ctx context.Context, name string) (out *Parameter, err error) {

	// setup tracing.
	newCtx, span := otel.Tracer(c.tracerName).Start(ctx, "Get")
	defer span.End()

	// retrieve parameter.
	param, err := c.GetMultiple(newCtx, name)
	if err != nil {
		return nil, err
	}
	return &param[0], nil
}

// GetMultiple retrieves one or more params from paramstore.
func (c *Client) GetMultiple(ctx context.Context, names ...string) (out Parameters, errs error) {

	// setup tracing.
	newCtx, span := otel.Tracer(c.tracerName).Start(ctx, "GetMultiple")
	defer span.End()

	// retrieve params in batches.
	var invalid []string
	for i := 0; i < len(names); i += c.batchSize {

		// determine rolling batch size.
		size := i + c.batchSize
		if size > len(names) {
			size = len(names)
		}

		// retrieve params.
		in := &ssm.GetParametersInput{
			Names:          names[i:size],
			WithDecryption: &c.withDecryption,
		}
		resp, err := c.ssmsvc.GetParameters(newCtx, in)
		if err != nil {
			c.logger.Error("failed to get parameters",
				"error", err,
				"names", in.Names,
				"decryption", *in.WithDecryption,
			)
			errs = multierror.Append(errs, err)
			continue
		}

		// parse params from response.
		for _, p := range resp.Parameters {

			// there's a very slim change the Value is missing.
			if p.Value == nil {
				p.Value = aws.String("")
			}

			out = append(out, Parameter{
				Name:  *p.Name,
				Value: *p.Value,
				Type:  ParameterType(p.Type),
			})
		}
		invalid = append(invalid, resp.InvalidParameters...)
	}

	// return params + errs.
	if len(invalid) > 0 {
		for _, i := range invalid {
			c.logger.Warn("found invalid parameters", "param", i)
			errs = multierror.Append(errs, fmt.Errorf("%q is an invalid parameter", i))
		}
	}
	return out, errs
}
