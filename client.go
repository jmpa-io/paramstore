package paramstore

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
)

const (
	Name = "paramstore"
)

// Client is the mechanism for someone to interact with this package.
type Client struct {

	// config.
	name string

	// clients.
	ssmsvc *ssm.Client

	// aws.
	awsRegion      string // the aws region to use when doing things with paramstore.
	batchSize      int    // batch size used when retrieving parameters.
	withDecryption bool   // decrypts the parameters when retrieving them.
	keyId          string // the KMS key to use when encrypting and decrypting parameters from paramstore.

	// misc.
	logger zerolog.Logger
}

// New returns a new Client.
func New(ctx context.Context, options ...Option) (*Client, error) {

	// setup tracing.
	newCtx, span := otel.Tracer(Name).Start(ctx, "New")
	defer span.End()

	// setup client w/ default values.
	c := &Client{
		name: Name,

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
