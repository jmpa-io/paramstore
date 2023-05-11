package paramstore

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
)

// Client is the mechanism for someone to interact with this package.
type Client struct {

	// config.
	name string

	// clients.
	ssmsvc *ssm.Client

	// aws.
	awsRegion      string
	batchSize      int
	withDecryption bool

	// misc.
	logger zerolog.Logger
}

// New returns a new Client.
func New(ctx context.Context, options ...Option) (*Client, error) {

	// setup tracing.
	name := "paramstore"
	newCtx, span := otel.Tracer(name).Start(ctx, "New")
	defer span.End()

	// setup client w/ default values.
	c := &Client{
		name: name,

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

	return c, nil
}
