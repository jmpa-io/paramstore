package paramstore

import (
	"fmt"

	"github.com/rs/zerolog"
)

// Option configures a paramstore client.
type Option func(*Client) error

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

// WithDecryption configures the decryption used by the client when retrieving
// from paramstore.
func WithDecryption(decryption bool) Option {
	return func(c *Client) error {
		c.withDecryption = decryption
		return nil
	}
}

// WithLogger configures the logger used in the client.
func WithLogger(logger zerolog.Logger) Option {
	return func(c *Client) error {
		c.logger = logger
		return nil
	}
}
