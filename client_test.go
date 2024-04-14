package paramstore

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

var (
	// debug logger.
	logger = zerolog.New(os.Stdout).
		With().Timestamp().
		Logger().Level(zerolog.DebugLevel)
)

func Test_New(t *testing.T) {
	tests := map[string]struct {
		options []Option
		want    *Client
		err     string
	}{
		"default": {
			want: &Client{
				awsRegion:      "ap-southeast-2",
				batchSize:      10,
				withDecryption: false,
				logger:         zerolog.Logger{},
			},
		},
		"set aws region": {
			options: []Option{WithAWSRegion("us-east-1")},
			want: &Client{
				awsRegion:      "us-east-1",
				batchSize:      10,
				withDecryption: false,
				logger:         zerolog.Logger{},
			},
		},
		"set batch size": {
			options: []Option{WithBatchSize(5)},
			want: &Client{
				awsRegion:      "ap-southeast-2",
				batchSize:      5,
				withDecryption: false,
				logger:         zerolog.Logger{},
			},
		},
		"set with decryption": {
			options: []Option{WithDecryption(true)},
			want: &Client{
				awsRegion:      "ap-southeast-2",
				batchSize:      10,
				withDecryption: true,
				logger:         zerolog.Logger{},
			},
		},
		"set with logger": {
			options: []Option{WithLogger(logger)},
			want: &Client{
				awsRegion:      "ap-southeast-2",
				batchSize:      10,
				withDecryption: false,
				logger:         logger,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := New(context.Background(), tt.options...)
			if tt.err != "" && err != nil {
				if !strings.Contains(err.Error(), tt.err) {
					t.Errorf("New() returned an unexpected error; want=%v, got=%v", tt.err, err)
				}
				return
			}
			if err != nil {
				t.Errorf("New() returned an error; error=%v", err)
				return
			}
			if got.awsRegion != tt.want.awsRegion ||
				got.batchSize != tt.want.batchSize ||
				got.withDecryption != tt.want.withDecryption ||
				got.logger.GetLevel() != tt.want.logger.GetLevel() {
				t.Errorf("New() returned unexpected configuration; want=%+v, got=%+v\n", tt.want, got)
				return
			}
		})
	}
}
