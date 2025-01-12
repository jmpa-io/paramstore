package paramstore

import "fmt"

// testParameter represents a Parameter, but used for testing.
type testParameter struct {
	Parameter
}

// testParameters is a slice of testParameter.
type testParameters []testParameter

// a helper function that converts testParameters to a slice of string.
func (parameters testParameters) toSliceString() (out []string) {
	for _, p := range parameters {
		out = append(out, p.Name)
	}
	return out
}

// a helper function that converts testParameters to a map[string]Parameter.
func (parameters testParameters) toMap() (out map[string]Parameter) {
	out = make(map[string]Parameter)
	for _, p := range parameters {
		out[p.Name] = p.Parameter
	}
	return out
}

// a helper function that converts testParameters to a Parameters type.
func (parameters testParameters) toParameters() (out Parameters) {
	for _, p := range parameters {
		out = append(out, p.Parameter)
	}
	return out
}

// a helper function to return a single Parameter from a slice of testParameters.
func (parameters testParameters) toParameter() (out *Parameter) {
	return &parameters.toParameters()[0]
}

// a helper function to return a slice of errors from a slice of testParameters.
// pass a formatted string as an arg to this function and the name of the
// parameter will be added to the error returned in the slice of errors.
func (parameters testParameters) toSliceError(format string) (out []error) {
	for _, p := range parameters {
		out = append(out, fmt.Errorf(format, p.Name))
	}
	return out
}

// a helper function to return a single formatted error from a slice of testParameters.
func (parameters testParameters) toError(format string) error {
	return fmt.Errorf(format, parameters.toParameter().Name)
}
