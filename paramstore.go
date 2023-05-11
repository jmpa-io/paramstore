package paramstore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go/aws"
	"go.opentelemetry.io/otel"
)

// Get retrieves one or more params from paramstore.
func (c *Client) Get(ctx context.Context, names ...string) (out []Param, errs int) {

	// setup tracing.
	newCtx, span := otel.Tracer(c.name).Start(ctx, "Get")
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
			c.logger.Error().
				Err(err).
				Strs("names", in.Names).
				Bool("decryption", *in.WithDecryption).
				Msg("failed to get parameters")
			errs++
			continue
		}

		// parse params from response.
		for _, p := range resp.Parameters {
			// there's a very slim change the Value is missing.
			if p.Value == nil {
				p.Value = aws.String("")
			}
			out = append(out, Param{
				Name:  *p.Name,
				Value: *p.Value,
				Type:  ParamType(p.Type),
			})
		}
		invalid = append(invalid, resp.InvalidParameters...)
	}

	// return params + errs.
	if len(invalid) > 0 {
		for _, i := range invalid {
			c.logger.Warn().
				Str("param", i).
				Msg("found invalid parameter")
			errs++
		}
	}
	return out, errs
}
