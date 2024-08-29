package paramstore

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/rs/zerolog"
)

// testParam represents a Param, but used for testing.
type testParam struct {
	Param
}

// testParams is a slice ot testParam.
type testParams []testParam

// a helper function that converts testParams to a slice of strings.
func (params testParams) convertToSliceString() (out []string) {
	for _, p := range params {
		out = append(out, p.Name)
	}
	return out
}

// a helper function that converts testParams to a map of Param.
func (params testParams) convertToMap() (out map[string]Param) {
	out = make(map[string]Param)
	for _, p := range params {
		out[p.Name] = p.Param
	}
	return out
}

// a helper function that converts testParams to a slice of Params.
func (params testParams) convertToParams() (out Params) {
	for _, p := range params {
		out = append(out, p.Param)
	}
	return out
}

// goodParams represents good parameters used around the tests in this file.
var goodParams = testParams{
	{
		Param{
			Name:  "/hello",
			Value: "this is (possibly) hidden",
			Type:  ParamType(types.ParameterTypeSecureString),
		},
	},
	{
		Param{
			Name:  "/world",
			Value: "this is plain text",
			Type:  ParamType(types.ParameterTypeString),
		},
	},
	{
		Param{
			Name:  "/test",
			Value: "this,is,a,comma,list",
			Type:  ParamType(types.ParameterTypeStringList),
		},
	},
}

// ssmGetter is a function used to retrieve parameters from AWS SSM Paramstore.
var ssmGetter = func(ctx context.Context, params *ssm.GetParametersInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersOutput, error) {
	if len(params.Names) == 0 {
		return nil, errors.New("no parameters specified")
	}
	out := &ssm.GetParametersOutput{
		Parameters: make([]types.Parameter, len(params.Names)),
	}
	m := goodParams.convertToMap()
	for i, n := range params.Names {
		param, found := m[n]
		if !found {
			return nil, fmt.Errorf("%q is not found in test parameters", n)
		}
		out.Parameters[i] = types.Parameter{
			Name:  &param.Name,
			Value: &param.Value,
			Type:  types.ParameterType(param.Type),
		}
	}
	return out, nil
}

// ssmPutter is a function used to upload parameters to AWS SSM Paramstore.
var ssmPutter = func(ctx context.Context, param *ssm.PutParameterInput, optFns ...func(*ssm.Options)) (*ssm.PutParameterOutput, error) {
	m := goodParams.convertToMap()
	_, found := m[*param.Name]
	if !found {
		return nil, fmt.Errorf("%q is not found in test parameters", *param.Name)
	}
	return &ssm.PutParameterOutput{
		Tier: param.Tier,
	}, nil
}

// ssmDeleter is a function used to delete parameters in AWS SSM Paramstore.
var ssmDeleter = func(ctx context.Context, params *ssm.DeleteParametersInput, optFns ...func(*ssm.Options)) (*ssm.DeleteParametersOutput, error) {
	if len(params.Names) == 0 {
		return nil, errors.New("no parameters specified")
	}
	out := &ssm.DeleteParametersOutput{
		DeletedParameters: make([]string, len(params.Names)),
	}
	m := goodParams.convertToMap()
	for i, n := range params.Names {
		param, found := m[n]
		if !found {
			return nil, fmt.Errorf("%q is not found in test parameters", n)
		}
		out.DeletedParameters[i] = param.Name
	}
	return out, nil
}

func Test_Get(t *testing.T) {

	// setup client.
	client := &Client{
		withDecryption: true,
		logger:         zerolog.Nop(),
		batchSize:      2,
		ssmsvc: &mockSSMClient{
			GetParametersFunc: ssmGetter,
		},
	}

	// tests.
	tests := map[string]struct {
		names []string
		want  []Param
		errs  int
	}{
		"retrieves parameters successfully": {
			names: goodParams.convertToSliceString(),
			want:  goodParams.convertToParams(),
			errs:  0,
		},
		// TODO retrieve with decryption set of false.
		// TODO failures.
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, errs := client.Get(context.Background(), tt.names...)
			if errs != tt.errs {
				t.Errorf("client.Get() returned a number of unexpected errors; want=%v, got=%v", tt.errs, errs)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("client.Get() returned more or less than expected params; want=%v got=%v", len(tt.want), len(got))
				return
			}
			for i, p := range got {
				if p.Name != tt.want[i].Name ||
					p.Value != tt.want[i].Value ||
					p.Type != tt.want[i].Type {
					t.Errorf("client.Get() returned values that don't match at index %v; want=%+v, got=%+v", i, tt.want[i], p)
					return
				}
			}
		})
	}
}

func Test_Put(t *testing.T) {

	// setup client.
	client := &Client{
		logger: zerolog.Nop(),
		ssmsvc: &mockSSMClient{
			PutParameterFunc: ssmPutter,
		},
	}

	// tests.
	tests := map[string]struct {
		params Params
		errs   int
	}{
		"puts parameters successfully": {
			params: goodParams.convertToParams(),
			errs:   0,
		},
		// TODO failures.
		// TODO test overwrite.
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			errs := client.Put(context.Background(), tt.params)
			if errs != tt.errs {
				t.Errorf("client.Put() returned a number of unexpected errors; want=%v, got=%v", tt.errs, errs)
				return
			}
		})
	}
}

func Test_Delete(t *testing.T) {

	// setup client.
	client := &Client{
		withDecryption: true,
		logger:         zerolog.Nop(),
		batchSize:      2,
		ssmsvc: &mockSSMClient{
			DeleteParametersFunc: ssmDeleter,
		},
	}

	// tests.
	tests := map[string]struct {
		names []string
		errs  int
	}{
		"deletes parameters successfully": {
			names: goodParams.convertToSliceString(),
			errs:  0,
		},
		// TODO failures.
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			errs := client.Delete(context.Background(), tt.names...)
			if errs != tt.errs {
				t.Errorf("client.Delete() returned a number of unexpected errors; want=%v, got=%v", tt.errs, errs)
				return
			}
		})
	}
}
