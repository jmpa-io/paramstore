package paramstore

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
)

var (
	// test logger.
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05"))
			}
			return a
		},
	}))
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
				logger:         slog.Default(),
			},
		},
		"with log level (debug)": {
			options: []Option{WithLogLevel(slog.LevelDebug)},
			want: &Client{
				awsRegion:      "ap-southeast-2",
				batchSize:      10,
				withDecryption: false,
				logger:         slog.Default(),
				logLevel:       -4,
			},
		},
		"with log level (info)": {
			options: []Option{WithLogLevel(slog.LevelInfo)},
			want: &Client{
				awsRegion:      "ap-southeast-2",
				batchSize:      10,
				withDecryption: false,
				logger:         slog.Default(),
				logLevel:       0,
			},
		},
		"with log level (warn)": {
			options: []Option{WithLogLevel(slog.LevelWarn)},
			want: &Client{
				awsRegion:      "ap-southeast-2",
				batchSize:      10,
				withDecryption: false,
				logger:         slog.Default(),
				logLevel:       4,
			},
		},
		"with log level (error)": {
			options: []Option{WithLogLevel(slog.LevelError)},
			want: &Client{
				awsRegion:      "ap-southeast-2",
				batchSize:      10,
				withDecryption: false,
				logger:         slog.Default(),
				logLevel:       8,
			},
		},
		"with logger": {
			options: []Option{WithLogger(logger)},
			want: &Client{
				awsRegion:      "ap-southeast-2",
				batchSize:      10,
				withDecryption: false,
				logger:         logger,
			},
		},
		"with aws region": {
			options: []Option{WithAWSRegion("us-east-1")},
			want: &Client{
				awsRegion:      "us-east-1",
				batchSize:      10,
				withDecryption: false,
				logger:         slog.Default(),
			},
		},
		"with batch size (n > min && n < max)": {
			options: []Option{WithBatchSize(5)},
			want: &Client{
				awsRegion:      "ap-southeast-2",
				batchSize:      5,
				withDecryption: false,
				logger:         slog.Default(),
			},
		},
		"with batch size (n < min)": {
			options: []Option{WithBatchSize(0)},
			want: &Client{
				awsRegion:      "ap-southeast-2",
				batchSize:      5,
				withDecryption: false,
				logger:         slog.Default(),
			},
			err: "batchSize must be greater than 0",
		},
		"with batch size (n > max)": {
			options: []Option{WithBatchSize(11)},
			want: &Client{
				awsRegion:      "ap-southeast-2",
				batchSize:      5,
				withDecryption: false,
				logger:         slog.Default(),
			},
			err: "batchSize must be less than or equal to 10",
		},
		"with decryption": {
			options: []Option{WithDecryption(true)},
			want: &Client{
				awsRegion:      "ap-southeast-2",
				batchSize:      10,
				withDecryption: true,
				logger:         slog.Default(),
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
			switch {
			case
				got.logLevel != tt.want.logLevel,
				(got.logger != slog.Default() && tt.want.logger != slog.Default()) && got.logger != tt.want.logger,
				got.awsRegion != tt.want.awsRegion,
				got.withDecryption != tt.want.withDecryption,
				got.batchSize != tt.want.batchSize:
				t.Errorf(
					"New() returned unexpected configuration; want=%+v, got=%+v\n",
					tt.want,
					got,
				)
				return
			}
		})
	}
}
