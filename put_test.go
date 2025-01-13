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

// mockPutParameter is a mock used to mimic the behavior of uploading a
// parameter to AWS SSM Parameter Store.
func mockPutParameter(
	t string,
) func(ctx context.Context, input *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error) {
	return func(ctx context.Context, input *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error) {
		out := &ssm.PutParameterOutput{}
		switch t {
		case "success":
			// do nothing.
		case "error":
			return nil, fmt.Errorf("failed to put parameter: %v", *input.Name)
		}
		return out, nil
	}
}

func Test_Put(t *testing.T) {
	tests := map[string]struct {
		client     *Client
		parameters Parameters
		errs       *multierror.Error
	}{
		"put parameter": {
			client: &Client{
				logger:    slog.Default(),
				batchSize: 1,
				ssmsvc: &mockSSMClient{
					PutParameterFunc: mockPutParameter("success"),
				},
			},
			parameters: validTestdata.toParameters(),
			errs:       nil,
		},
		"puts parameters": {
			client: &Client{
				logger:    slog.Default(),
				batchSize: 1,
				ssmsvc: &mockSSMClient{
					PutParameterFunc: mockPutParameter("success"),
				},
			},
			parameters: validTestdata.toParameters(),
			errs:       nil,
		},
		"catch fail to put parameters": {
			client: &Client{
				logger: slog.Default(),
				ssmsvc: &mockSSMClient{
					PutParameterFunc: mockPutParameter("error"),
				},
			},
			parameters: validTestdata.toParameters(),
			errs: &multierror.Error{
				Errors: validTestdata.toSliceError("failed to put parameter: %v"),
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.client.Put(context.Background(), tt.parameters)

			// determine what to do if the error returned is a multierror or not.
			// NOTE: got is populated by the errors.As() function.
			var errs *multierror.Error
			if errors.As(err, &errs) {

				// check if the number of errors match with the expected errors.
				if len(errs.Errors) != len(tt.errs.Errors) {
					t.Errorf(
						"Put() returned unexpected number of errors;\nwant=%v\ngot=%v\n",
						len(tt.errs.Errors),
						len(errs.Errors),
					)
					return
				}

				// compare each error returned with the expected errors.
				for i, expectedErr := range tt.errs.Errors {
					if errs.Errors[i].Error() != expectedErr.Error() {
						t.Errorf(
							"Put() got unexpected error;\nwant=%v\ngot=%v\n",
							expectedErr,
							errs.Errors[i].Error(),
						)
						return
					}
				}

			} else if tt.errs != nil {
				// an error was returned but it wasn't a multierror.
				t.Errorf("Put() expected a multierror but got a regular error instead;\nwant=%v\ngot=%v\n", tt.errs, err)
				return
			}
		})
	}
}
