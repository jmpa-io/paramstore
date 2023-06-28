package paramstore

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

func Test_Get(t *testing.T) {

	// setup client.
	client := &Client{
		withDecryption: true,
		logger:         logger.With().Logger(),
		ssmsvc: &mockSSMClient{
			GetParametersFunc: func(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error) {
				if len(params.Names) == 0 {
					return nil, errors.New("no parameters specified")
				}
				var out *ssm.GetParametersOutput
				for i, n := range params.Names {
					out.Parameters = append(out.Parameters, types.Parameter{
						Name:  aws.String(n),
						Value: aws.String(fmt.Sprintf("%v", i)),
						Type:  types.ParameterTypeSecureString,
					})
				}
				return out, nil
			},
		},
	}

	// tests.
	tests := map[string]struct {
		names []string
		want  []Param
		errs  int
	}{
		"retrieves parameters successfully": {
			names: []string{
				"/hello",
				"/world",
			},
			want: []Param{
				{
					Name:  "/hello",
					Value: "1",
					Type:  ParamType(types.ParameterTypeSecureString),
				},
				{
					Name:  "/world",
					Value: "2",
					Type:  ParamType(types.ParameterTypeSecureString),
				},
			},
		},
		// {
		// 	name:     "Error: No parameters specified",
		// 	input:    []string{},
		// 	expected: []Param{},
		// },
		// {
		// 	name:     "Error: Failed to retrieve parameters",
		// 	input:    []string{"param1"},
		// 	expected: []Param{},
		// },
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, errs := client.Get(context.Background(), tt.names...)
			if errs > 0 {
				t.Errorf("Expected %d invalid parameters, got %d", tt.errs, errs)
				return
			}
			fmt.Printf("%+v\n", got)

			// // Verify the number of parameters and invalid parameters
			// if len(out) != len(tt.expected) {
			// 	t.Errorf("Expected %d parameters, got %d", len(tt.expected), len(out))
			// }
			// if errs != tt.errCount {
			// }

			// // Verify the parameters individually
			// for i, p := range out {
			// 	if !reflect.DeepEqual(p, tt.expected[i]) {
			// 		t.Errorf("Parameter mismatch at index %d. Expected %+v, got %+v", i, tt.expected[i], p)
			// 	}
			// }
		})
	}
}

// mockSSMClient is a mock implementation of the ssm.Client.
type mockSSMClient struct {

	// clients.
	ssm.Client

	// funcs.
	GetParametersFunc    func(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error)
	PutParameterFunc     func(ctx context.Context, params *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error)
	DeleteParametersFunc func(ctx context.Context, params *ssm.DeleteParametersInput, optFns ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error)
}

// GetParameters mocks the GetParameters function.
func (m *mockSSMClient) GetParameters(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error) {
	if m.GetParametersFunc != nil {
		return m.GetParametersFunc(ctx, params, optFns...)
	}
	return nil, errors.New("GetParametersFunc is not implemented")
}

// PutParameter mocks the PutParameter function.
func (m *mockSSMClient) PutParameter(ctx context.Context, params *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error) {
	if m.PutParameterFunc != nil {
		return m.PutParameterFunc(ctx, params, optFns...)
	}
	return nil, errors.New("PutParameterFunc is not implemented")
}

// DeleteParameters mocks the DeleteParameters function.
func (m *mockSSMClient) DeleteParameters(ctx context.Context, params *ssm.DeleteParametersInput, optFns ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error) {
	if m.DeleteParametersFunc != nil {
		return m.DeleteParametersFunc(ctx, params, optFns...)
	}
	return nil, errors.New("DeleteParametersFunc is not implemented")
}
