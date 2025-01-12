package paramstore

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/hashicorp/go-multierror"
)

// ssmPutter is a test function used to mock uploading parameters to AWS SSM Parameter Store.
var ssmPutter = func(ctx context.Context, input *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error) {
	if len(*input.Name) == 0 {
		return nil, errors.New("no parameter specified")
	}
	testdata := validTestdata.toMap()
	_, found := testdata[*input.Name]
	if !found {
		return nil, fmt.Errorf("%q is not found in testdata", *input.Name)
	}
	return &ssm.PutParameterOutput{
		Tier: input.Tier,
	}, nil
}

func Test_Put(t *testing.T) {
	tests := map[string]struct {
		client *Client
		params Parameters
		errs   *multierror.Error
	}{
		"put parameter": {
			client: &Client{
				logger: slog.Default(),
				ssmsvc: &mockSSMClient{
					PutParameterFunc: ssmPutter,
				},
			},
			params: validTestdata.toParameters(),
			errs:   nil,
		},
		"puts parameters": {
			client: &Client{
				logger: slog.Default(),
				ssmsvc: &mockSSMClient{
					PutParameterFunc: ssmPutter,
				},
			},
			params: validTestdata.toParameters(),
			errs:   nil,
		},
		// TODO failures.
		// TODO test overwrite.
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.client.Put(context.Background(), tt.params)

			var errs *multierror.Error
			if !errors.As(err, &errs) {
				// expected multi-error.
			}

			if errs != tt.errs {
				t.Errorf(
					"Put() returned a number of unexpected errors; want=%v, got=%v",
					tt.errs,
					errs,
				)
				return
			}
		})
	}
}
