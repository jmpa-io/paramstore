package paramstore

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
)

// iSSMClient is an interface for ssm.Client.
type iSSMClient interface {
	GetParameters(
		ctx context.Context,
		params *ssm.GetParametersInput,
		optFns ...func(*ssm.Options),
	) (*ssm.GetParametersOutput, error)
	PutParameter(
		ctx context.Context,
		params *ssm.PutParameterInput,
		optFns ...func(*ssm.Options),
	) (*ssm.PutParameterOutput, error)
	DeleteParameters(
		ctx context.Context,
		params *ssm.DeleteParametersInput,
		optFns ...func(*ssm.Options),
	) (*ssm.DeleteParametersOutput, error)
}

// Client is the mechanism for someone to interact with this package.
type Client struct {

	// tracing.
	tracerName string // The name of the tracer output in the traces.

	// clients.
	ssmsvc iSSMClient

	// aws.
	awsRegion      string // The aws region to use when doing things with paramstore.
	batchSize      int    // The batch size used when retrieving parameters.
	withDecryption bool   // This decrypts parameters when retrieving them.
	keyId          string // The KMS key to use when encrypting and decrypting parameters from paramstore.

	// misc.
	logger zerolog.Logger
}

// New returns a new Client.
func New(ctx context.Context, options ...Option) (*Client, error) {

	// setup tracing.
	tracerName := "paramstore"
	newCtx, span := otel.Tracer(tracerName).Start(ctx, "New")
	defer span.End()

	// setup client w/ default values.
	c := &Client{
		tracerName: tracerName,

		awsRegion:      "ap-southeast-2",
		batchSize:      10,
		withDecryption: false,
	}

	// overwrite client with given options.
	for _, o := range options {
		if err := o(c); err != nil {
			return nil, fmt.Errorf("failed to setup client: %v", err)
		}
	}

	// load aws config.
	cfg, err := config.LoadDefaultConfig(newCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %v", err)
	}

	// setup ssm client.
	c.ssmsvc = ssm.New(ssm.Options{
		Region:      c.awsRegion,
		Credentials: cfg.Credentials,
	})

	c.logger.Debug().Msg("client setup successfully")
	return c, nil
}
