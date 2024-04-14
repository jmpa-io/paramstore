package paramstore

import (
	"context"

	"go.opentelemetry.io/otel"
)

// Param is a thin wrapper over ssm.Parameter.
type Param struct {
	Name      string    // the name of the parameter.
	Value     string    // the value of the parameter.
	Type      ParamType // the type of the parameter.
	Overwrite bool      // overwrite existing parameters during Put().
}

// Params is a slice of Param.
type Params []Param

// NamesToSliceString converts a Params slice of names into a slice string.
func (params Params) NamesToSliceString(ctx context.Context) (out []string) {

	// setup tracing.
	_, span := otel.Tracer(Name).Start(ctx, "NamesToSliceString")
	defer span.End()

	// extract names.
	for _, p := range params {
		out = append(out, p.Name)
	}
	return out
}

// ParamType is a thin wrapper over aws types.ParamType.
type ParamType string

const (
	StringParamType       ParamType = "String"
	StringListParamType   ParamType = "StringList"
	SecureStringParamType ParamType = "SecureString"
)
