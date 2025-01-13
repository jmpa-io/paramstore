package paramstore

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/go-multierror"
)

// mockGetParameters is a mock used to mimic the behavior of pulling multiple
// parameters from AWS SSM Parameter Store.
func mockGetParameters(
	t string,
) func(ctx context.Context, input *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error) {
	return func(ctx context.Context, input *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error) {
		out := &ssm.GetParametersOutput{}
		switch t {
		case "success":
			out.Parameters = make([]types.Parameter, len(input.Names))
			testdata := validTestdata.toMap()
			for i, n := range input.Names {
				param, found := testdata[n]
				if !found {
					return nil, fmt.Errorf("%q is not found in testdata", n)
				}
				out.Parameters[i] = types.Parameter{
					Name:  &param.Name,
					Value: &param.Value,
					Type:  types.ParameterType(param.Type),
				}
			}
		case "invalid":
			for _, n := range input.Names {
				out.InvalidParameters = append(out.InvalidParameters, n)
			}
		case "error":
			return nil, fmt.Errorf(
				"failed to get parameter: %v",
				strings.Join(input.Names, ", "),
			)
		}
		return out, nil
	}
}

func Test_Get(t *testing.T) {
	tests := map[string]struct {
		client *Client
		name   string
		want   *Parameter
		errs   *multierror.Error
	}{
		"get parameter": {
			client: &Client{
				withDecryption: true,
				logger:         slog.Default(),
				batchSize:      1,
				ssmsvc: &mockSSMClient{
					GetParametersFunc: mockGetParameters("success"),
				},
			},
			name: validTestdata.toParameter().Name,
			want: validTestdata.toParameter(),
		},
		"catch invalid parameters": {
			client: &Client{
				logger:    slog.Default(),
				batchSize: 1,
				ssmsvc: &mockSSMClient{
					GetParametersFunc: mockGetParameters("invalid"),
				},
			},
			name: validTestdata.toParameter().Name,
			errs: &multierror.Error{
				Errors: []error{validTestdata.toError("%q is an invalid parameter")},
			},
		},
		"catch fail to get parameter": {
			client: &Client{
				logger:    slog.Default(),
				batchSize: 1,
				ssmsvc: &mockSSMClient{
					GetParametersFunc: mockGetParameters("error"),
				},
			},
			name: validTestdata.toParameter().Name,
			errs: &multierror.Error{
				Errors: []error{validTestdata.toError("failed to get parameter: %v")},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tt.client.Get(context.Background(), tt.name)

			// determine what to do if the error returned is a multierror or not.
			// NOTE: got is populated by the errors.As() function.
			var errs *multierror.Error
			if errors.As(err, &errs) {

				// check if the number of errors match with the expected errors.
				if len(errs.Errors) != len(tt.errs.Errors) {
					t.Errorf(
						"Get() returned unexpected number of errors;\nwant=%v\ngot=%v\n",
						tt.errs.Errors,
						errs.Errors,
					)
					return
				}

				// compare each error returned with the expected errors.
				for i, expectedErr := range tt.errs.Errors {
					if errs.Errors[i].Error() != expectedErr.Error() {
						t.Errorf(
							"Get() got unexpected error;\nwant=%v\ngot=%v\n",
							expectedErr,
							errs.Errors[i].Error(),
						)
						return
					}
				}

			} else if tt.errs != nil {
				// an error was returned but it wasn't a multierror.
				t.Errorf("Get() expected a multierror but got a regular error instead;\nwant=%v\ngot=%v\n", tt.errs, err)
				return
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf(
					"Get() returned unexpected configuration;\nwant=%+v\ngot=%+v\n",
					tt.want,
					got,
				)
				return
			}
		})
	}
}

func Test_GetMultiple(t *testing.T) {
	tests := map[string]struct {
		client *Client
		names  []string
		want   Parameters
		errs   *multierror.Error
	}{
		"get parameters": {
			client: &Client{
				withDecryption: true,
				logger:         slog.Default(),
				batchSize:      2,
				ssmsvc: &mockSSMClient{
					GetParametersFunc: mockGetParameters("success"),
				},
			},
			names: validTestdata.toSliceString(),
			want:  validTestdata.toParameters(),
			errs:  nil,
		},
		"catch invalid parameters": {
			client: &Client{
				logger:    slog.Default(),
				batchSize: 1,
				ssmsvc: &mockSSMClient{
					GetParametersFunc: mockGetParameters("invalid"),
				},
			},
			names: validTestdata.toSliceString(),
			errs: &multierror.Error{
				Errors: validTestdata.toSliceError("%q is an invalid parameter"),
			},
		},
		"catch fail to get parameters": {
			client: &Client{
				logger:    slog.Default(),
				batchSize: 1,
				ssmsvc: &mockSSMClient{
					GetParametersFunc: mockGetParameters("error"),
				},
			},
			names: validTestdata.toSliceString(),
			errs: &multierror.Error{
				Errors: validTestdata.toSliceError("failed to get parameter: %v"),
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := tt.client.GetMultiple(context.Background(), tt.names...)

			// determine what to do if the error returned is a multierror or not.
			// NOTE: got is populated by the errors.As() function.
			var errs *multierror.Error
			if errors.As(err, &errs) {

				// check if the number of errors match with the expected errors.
				if len(errs.Errors) != len(tt.errs.Errors) {
					t.Errorf(
						"GetMultiple() returned unexpected number of errors;\nwant=%v\ngot=%v\n",
						tt.errs.Errors,
						errs.Errors,
					)
					return
				}

				// compare each error returned with the expected errors.
				for i, expectedErr := range tt.errs.Errors {
					if errs.Errors[i].Error() != expectedErr.Error() {
						t.Errorf(
							"GetMultiple() got unexpected error;\nwant=%v\ngot=%v\n",
							expectedErr,
							errs.Errors[i].Error(),
						)
						return
					}
				}

			} else if tt.errs != nil {
				// an error was returned but it wasn't a multierror.
				t.Errorf("GetMultiple() expected a multierror but got a regular error instead;\nwant=%v\ngot=%v\n", tt.errs, err)
				return
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf(
					"GetMultiple() returned unexpected configuration;\nwant=%+v\ngot=%+v\n",
					tt.want,
					got,
				)
				return
			}
		})
	}
}
