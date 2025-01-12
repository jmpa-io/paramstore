package paramstore

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
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

// Client defines a client for this package.
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
	logLevel slog.Level   // The log level of the default logger.
	logger   *slog.Logger // The logger used in this client (custom or default).
}

// New creates and returns a new Client. The client itself is set up with
// tracing & logging. Additional options can be provided to modify its
// behavior, via the options slice. The client is used for interacting with
// parameters in AWS SSM Parameter Store.
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

	// overwrite client with any given options.
	for _, o := range options {
		if err := o(c); err != nil {
			return nil, ErrClientFailedToSetOption{err}
		}
	}

	// determine if the default logger should be used.
	if c.logger == nil {

		// use default logger.
		c.logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: c.logLevel, // default log level is 'INFO'.
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05"))
				}
				return a
			},
		}))

	}

	// load aws config.
	cfg, err := config.LoadDefaultConfig(newCtx)
	if err != nil {
		return nil, ErrClientFailedToLoadAWSConfig{err}
	}

	// setup ssm client.
	c.ssmsvc = ssm.New(ssm.Options{
		Region:      c.awsRegion,
		Credentials: cfg.Credentials,
	})

	c.logger.Debug("client setup successfully")
	return c, nil
}
