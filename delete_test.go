package paramstore

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/hashicorp/go-multierror"
)

// mockDeleteParameters is a mock used to mimic the behavior of deleting
// parameters from AWS SSM Parameter Store.
func mockDeleteParameters(
	t string,
) func(ctx context.Context, input *ssm.DeleteParametersInput, optFns ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error) {
	return func(ctx context.Context, input *ssm.DeleteParametersInput, optFns ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error) {
		out := &ssm.DeleteParametersOutput{}
		switch t {
		case "success":
			out.DeletedParameters = append(out.DeletedParameters, input.Names...)
		case "invalid":
			out.InvalidParameters = append(out.InvalidParameters, input.Names...)
		case "error":
			return nil, fmt.Errorf(
				"failed to delete parameter: %v",
				strings.Join(input.Names, ", "),
			)
		}
		return out, nil
	}
}

func Test_Delete(t *testing.T) {
	tests := map[string]struct {
		client *Client
		names  []string
		errs   *multierror.Error
	}{
		"delete parameters": {
			client: &Client{
				logger:    slog.Default(),
				batchSize: 1,
				ssmsvc: &mockSSMClient{
					DeleteParametersFunc: mockDeleteParameters("success"),
				},
			},
			names: validTestdata.toSliceString(),
		},
		"catch invalid parameters": {
			client: &Client{
				logger:    slog.Default(),
				batchSize: 10,
				ssmsvc: &mockSSMClient{
					DeleteParametersFunc: mockDeleteParameters("invalid"),
				},
			},
			names: validTestdata.toSliceString(),
			errs: &multierror.Error{
				Errors: validTestdata.toSliceError("%q is an invalid parameter"),
			},
		},
		"catch fail to delete parameters": {
			client: &Client{
				logger:    slog.Default(),
				batchSize: 1,
				ssmsvc: &mockSSMClient{
					DeleteParametersFunc: mockDeleteParameters("error"),
				},
			},
			names: validTestdata.toSliceString(),
			errs: &multierror.Error{
				Errors: validTestdata.toSliceError("failed to delete parameter: %v"),
			},
		},
		"catch no parameters given": {
			client: &Client{
				logger:    slog.Default(),
				batchSize: 1,
				ssmsvc: &mockSSMClient{
					DeleteParametersFunc: mockDeleteParameters("success"),
				},
			},
			names: []string{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.client.Delete(context.Background(), tt.names...)

			// determine what to do if the error returned is a multierror or not.
			// NOTE: got is populated by the errors.As() function.
			var errs *multierror.Error
			if errors.As(err, &errs) {

				// check if the number of errors match with the expected errors.
				if len(errs.Errors) != len(tt.errs.Errors) {
					t.Errorf(
						"Delete() returned unexpected number of errors;\nwant=%v\ngot=%v\n",
						len(tt.errs.Errors),
						len(errs.Errors),
					)
					return
				}

				// compare each error returned with the expected errors.
				for i, expectedErr := range tt.errs.Errors {
					if errs.Errors[i].Error() != expectedErr.Error() {
						t.Errorf(
							"Delete() got unexpected error;\nwant=%v\ngot=%v\n",
							expectedErr,
							errs.Errors[i].Error(),
						)
						return
					}
				}

			} else if tt.errs != nil {
				// an error was returned but it wasn't a multierror.
				t.Errorf("Delete() expected a multierror but got a regular error instead;\nwant=%v\ngot=%v\n", tt.errs, err)
				return
			}
		})
	}
}
