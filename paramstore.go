package paramstore

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type handler struct {
	ssmsvc *ssm.SSM
}

// sets up the handler used in public functions in this package.
func defaultHandler() (*handler, error) {

	// which region?
	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = os.Getenv("AWS_REGION")
	}
	if region == "" {
		region = "ap-southeast-2"
	}

	// setup session.
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(region)},
	})
	if err != nil {
		return nil, Error{fmt.Sprintf("failed to setup aws session: %v", err)}
	}

	// return handler.
	return &handler{
		ssmsvc: ssm.New(sess, nil),
	}, nil
}

// Param represents a param stored in paramstore.
type Param struct {
	Name  string
	Value string
	Type  string
}

type Params []Param

// retrieves params for the given paths from paramstore.
func (h *handler) Get(paths ...string) (Params, error) {

	// check input.
	if len(paths) == 0 {
		return nil, Error{"missing paths"}
	}

	// read params.
	var params Params
	var invalid []*string
	size := 10 // size of chunk.
	for b := 0; b < len(paths); b += size {
		n := b + size // number of params to read at a time.
		if n > len(paths) {
			n = len(paths)
		}

		// retrieve params.
		out, err := h.ssmsvc.GetParameters(&ssm.GetParametersInput{
			Names:          aws.StringSlice(paths[b:n]),
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			return nil, Error{fmt.Sprintf("failed to read params from paramstore: %v", err)}
		}

		// parse params.
		for _, p := range out.Parameters {
			if p.Value == nil { // slim chance the Value could missing.
				p.Value = aws.String("")
			}
			params = append(params, Param{
				Name:  *p.Name,
				Value: *p.Value,
				Type:  *p.Type,
			})
		}
		invalid = append(invalid, out.InvalidParameters...)
	}

	// any invalid params?
	if len(invalid) > 0 {
		return params, missingErr{invalid}
	}
	return params, nil
}

// Get retrieves a single param from paramstore, for the given path.
func Get(path string) (string, error) {

	// setup handler.
	h, err := defaultHandler()
	if err != nil {
		return "", Error{fmt.Sprintf("failed to default handler: %v", err)}
	}

	// retrieve param.
	params, err := h.Get(path)
	if err != nil {
		return "", err
	}

	// multiple params returned?
	if len(params) != 1 {
		return "", Error{fmt.Sprintf("%s matched %d parameters", path, len(params))}
	}
	return params[0].Value, nil
}

// GetPaths retrieves multiple params from paramstore at a time, for the given paths.
func GetPaths(paths ...string) (Params, error) {

	// setup handler.
	h, err := defaultHandler()
	if err != nil {
		return nil, Error{fmt.Sprintf("failed to default handler: %v", err)}
	}

	// retrieve + return params.
	return h.Get(paths...)
}

// uploads the given param to the given path.
func Put(params ...Param) error {

	// check input.
	if len(params) == 0 {
		return Error{"missing params"}
	}

	// setup handler.
	h, err := defaultHandler()
	if err != nil {
		return Error{fmt.Sprintf("failed to default handler: %v", err)}
	}

	// put params.
	for _, p := range params {
		_, err := h.ssmsvc.PutParameter(&ssm.PutParameterInput{
			Name:      aws.String(p.Name),
			Value:     aws.String(p.Value),
			Type:      aws.String(p.Type),
			Overwrite: aws.Bool(true),
		})
		if err != nil {
			return Error{fmt.Sprintf("failed to put parameter: %v", err)}
		}
	}
	return nil
}
