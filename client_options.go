package paramstore

import (
	"fmt"
	"log/slog"
)

// Option configures a paramstore client.
type Option func(*Client) error

// WithLogLevel sets the log level for the default logger.
func WithLogLevel(level slog.Level) Option {
	return func(c *Client) error {
		c.logLevel = level
		return nil
	}
}

// WithLogger configures the logger used in the client.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) error {
		c.logger = logger
		return nil
	}
}

// WithDecryption configures the decryption used by the client when retrieving
// from AWS SSM Parameter Store. This option must be given to decrypt any
// parameters returned to this client.
func WithDecryption(decryption bool) Option {
	return func(c *Client) error {
		c.withDecryption = decryption
		return nil
	}
}

// WithAWSRegion configures the AWS region used in the client.
func WithAWSRegion(region string) Option {
	return func(c *Client) error {
		c.awsRegion = region
		return nil
	}
}

const (
	// the min batch size used when uploading to paramstore.
	minBatchSize = 0

	// the max batch size used when uploading to paramstore.
	maxBatchSize = 10
)

// WithBatchSize configures the batchSize used by the client when retrieving
// from or uploading to paramstore.
func WithBatchSize(size int) Option {
	return func(c *Client) error {
		if size <= minBatchSize {
			return fmt.Errorf("batchSize must be greater than %v", minBatchSize)
		}
		if size > maxBatchSize {
			return fmt.Errorf("batchSize must be less than or equal to %v", maxBatchSize)
		}
		c.batchSize = size
		return nil
	}
}
