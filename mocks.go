package paramstore

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

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
