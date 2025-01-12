package paramstore

import "github.com/aws/aws-sdk-go-v2/service/ssm/types"

// ParameterType is a thin wrapper over ssm/types.ParameterType.
// NOTE:
// This type exists to reduce the number of packages that anyone using this
// package needs to explicitly import into their own programs.
type ParameterType types.ParameterType

const (
	ParameterTypeString       ParameterType = ParameterType(types.ParameterTypeString)
	ParameterTypeStringList   ParameterType = ParameterType(types.ParameterTypeStringList)
	ParameterTypeSecureString ParameterType = ParameterType(types.ParameterTypeSecureString)
)

// Parameter is a thin wrapper over ssm/types.Parameter.
type Parameter struct {
	Name      string        // The name of the parameter.
	Value     string        // The value of the parameter.
	Type      ParameterType // The type of the parameter.
	Overwrite bool          // Used to overwrite existing parameters during Put().
}

// Parameters is a slice of Parameter.
type Parameters []Parameter

// ToSliceString converts a Params slice of names into a slice string.
func (parameters Parameters) ToSliceString() (out []string) {
	for _, p := range parameters {
		out = append(out, p.Name)
	}
	return out
}
