package paramstore

// Param is a thin wrapper over ssm.Parameter.
type Param struct {
	Name      string    // the name of the parameter.
	Value     string    // the value of the parameter.
	Type      ParamType // the type of the parameter.
	Overwrite bool      // overwrite existing parameters during Put().
}

// ParamType is a thin wrapper over aws types.ParamType.
type ParamType string

const (
	ParamTypeString       ParamType = "String"
	ParamTypeStringList   ParamType = "StringList"
	ParamTypeSecureString ParamType = "SecureString"
)
