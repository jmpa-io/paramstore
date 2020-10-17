package paramstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// Handle abstracts the SSM service and Parameter Store functions.
type Handle struct {
	*ssm.Client
	KeyID string // Optional KMS key ID to use for encryption and decryption
}

// DefaultHandle returns a Handle with defaults set, obvs.
func DefaultHandle() (*Handle, error) {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %v", err)
	}
	reg := os.Getenv("AWS_DEFAULT_REGION")
	if reg == "" {
		reg = os.Getenv("AWS_REGION")
	}
	if reg == "" {
		reg = "ap-southeast-2"
	}
	cfg.Region = reg
	return &Handle{Client: ssm.New(ssm.Options{Credentials: cfg.Credentials, Region: cfg.Region})}, nil
}

// Param is a thin wrapper over ssm.Parameter.
type Param struct {
	Name      string // the name of the parameter
	Value     string // the value of the parameter
	Type      types.ParameterType
	Overwrite bool // overwrite existing parameters during Put()
}

// Exists checks if a single parameter exists in SSM.
func (h Handle) Exists(n string) (bool, error) {
	in := &ssm.GetParametersInput{
		Names: []*string{aws.String(n)},
	}
	out, err := h.GetParameters(context.Background(), in)
	if err != nil {
		return false, err
	}
	if len(out.InvalidParameters) > 0 {
		return false, nil
	}
	return true, nil
}

// Get accepts a number of parameter names and returns a slice of Param structs
// to match. If any parameter names are invalid (do not exist), an error is
// returned in addition to a slice of Param structs (if possible). Querying is
// performed in chunks to avoid paging.
func (h Handle) Get(n ...*string) ([]Param, error) {
	var pp []Param
	var invalid []*string

	for i := 0; i < len(n); i += 10 { // chunked because AWS
		x := i + 10
		if x > len(n) {
			x = len(n)
		}
		in := &ssm.GetParametersInput{
			Names:          n[i:x],
			WithDecryption: aws.Bool(true),
		}
		out, err := h.GetParameters(context.Background(), in)
		if err != nil {
			return nil, err
		}
		for _, p := range out.Parameters {
			// there's a very slim chance the Value could be missing.
			if p.Value == nil {
				p.Value = aws.String("")
			}
			pp = append(pp, Param{
				Name:  *p.Name,
				Value: *p.Value,
				Type:  p.Type,
			})
		}
		invalid = append(invalid, out.InvalidParameters...)
	}
	if len(invalid) > 0 {
		return pp, missingErr{invalid}
	}
	return pp, nil
}

// GetPath accepts a path and returns all parameters which match and an error.
// The second parameter, if true, will recurse results.
func (h Handle) GetPath(path string, rec bool) ([]Param, error) {
	var pp []Param
	if len(path) == 0 {
		return pp, errors.New("path must one or more chars")
	}
	in := &ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		Recursive:      aws.Bool(rec),
		WithDecryption: aws.Bool(true),
	}
	for {
		out, err := h.GetParametersByPath(context.Background(), in)
		if err != nil {
			return nil, err
		}
		for _, p := range out.Parameters {
			// there's a very slim chance the Value could be missing.
			if p.Value == nil {
				p.Value = aws.String("")
			}
			pp = append(pp, Param{
				Name:  *p.Name,
				Value: *p.Value,
				Type:  p.Type,
			})
		}
		if out.NextToken != nil {
			in.NextToken = out.NextToken
		}
	}
	return pp, nil
}

// Glob applies filepath globbing to all parameters and returns those that match.
func (h Handle) Glob(glob string) ([]Param, error) {
	ps, err := h.GetPath("/", true)
	if err != nil {
		return []Param{}, err
	}
	ms := []Param{}
	for _, p := range ps {
		m, err := filepath.Match(glob, p.Name)
		switch {
		case err != nil:
			return []Param{}, err
		case m:
			ms = append(ms, p)
		}
	}
	return ms, nil
}

// Decode fetches params into a struct using "ssm" struct tags.
// Nested structs are passed the tag of the parent as a prefix.
// Passing anything other than a pointer to a struct is guaranteed to end badly.
func (h Handle) Decode(pfx string, v interface{}) error {
	u := reflect.ValueOf(v)
	if u.Kind() == reflect.Ptr {
		u = u.Elem()
	}
	t := u.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		k := f.Tag.Get("ssm")
		switch {
		case f.Type.Kind() == reflect.Struct:
			if err := h.Decode(pfx+k, u.Field(i).Addr().Interface()); err != nil {
				return err
			}
		case k == "":
		case f.Type.Kind() == reflect.String:
			ps, err := h.Get(aws.String(pfx + k))
			if err != nil {
				return fmt.Errorf("failed to decode %s into %s: %s", pfx+k, f.Name, err)
			}
			u.Field(i).Set(reflect.ValueOf(ps[0].Value))
		}
	}
	return nil
}

// Encode puts params from a struct using "ssm" struct tags.
// Nested structs use the tag of the parent as a prefix.
// Passing anything other than a pointer to a struct is guaranteed to end badly.
func (h Handle) Encode(pfx string, v interface{}) error {
	u := reflect.ValueOf(v)
	if u.Kind() == reflect.Ptr {
		u = u.Elem()
	}
	t := u.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		k := f.Tag.Get("ssm")
		switch {
		case f.Type.Kind() == reflect.Struct:
			if err := h.Encode(pfx+k, u.Field(i).Interface()); err != nil {
				return err
			}
		case k == "":
		case f.Type.Kind() == reflect.String:
			p := Param{Name: pfx + k, Value: u.Field(i).String(), Type: "String", Overwrite: true}
			if err := h.Put(p); err != nil {
				return fmt.Errorf("failed to encode %s into %s: %s", f.Name, pfx+k, err)
			}
		}
	}
	return nil
}

// Put accepts a number of Param structs and writes them to Parameter Store in
// the other they are supplied. Paramters which already exist will cause an
// error to be returned immediately unless the Overwrite field is set to true.
func (h Handle) Put(pp ...Param) error {
	for _, p := range pp {
		i := &ssm.PutParameterInput{
			Name:      aws.String(p.Name),
			Value:     aws.String(p.Value),
			Type:      p.Type,
			Overwrite: aws.Bool(p.Overwrite),
		}
		if h.KeyID != "" {
			i.KeyId = aws.String(h.KeyID)
		}
		_, err := h.PutParameter(context.Background(), i)
		if err != nil {
			return err // TODO discuss
		}
	}
	return nil
}

// Del accepts a number of parameter names and deletes them.
func (h Handle) Del(n ...*string) error {
	// TODO test if 10-param limit is in effect here
	in := &ssm.DeleteParametersInput{
		Names: n,
	}
	out, err := h.DeleteParameters(context.Background(), in)
	if err != nil {
		return err
	}
	if len(out.InvalidParameters) > 0 {
		return missingErr{out.InvalidParameters}
	}
	return nil
}

func (p Param) String() string {
	v := p.Value
	if p.Type == "SecureString" {
		v = "***"
	}
	return fmt.Sprintf("%s\t%v\t%s", p.Name, v, p.Type)
}

// Get gets a single parameter using the default AWS config. The AWS region, if not present,
// will default to ap-southeast-2. The parameter will be decrypted and its value is returned.
func Get(n *string) (string, error) {
	h, err := DefaultHandle()
	if err != nil {
		return "", err
	}
	ps, err := h.Get(n)
	if err != nil {
		return "", err
	}
	if len(ps) != 1 {
		return "", fmt.Errorf("%s: %d parameters matched", *n, len(ps))
	}
	return ps[0].Value, nil
}

// Put puts a single parameter using the default handle.
func Put(p Param) error {
	h, err := DefaultHandle()
	if err != nil {
		return err
	}
	return h.Put(p)
}

// Del deletes n number of parameters using the default handle.
func Del(p ...*string) error {
	h, err := DefaultHandle()
	if err != nil {
		return err
	}
	return h.Del(p...)
}

// Decode fetches params into a struct using "ssm" struct tags.
// Nested structs are passed the tag of the parent as a prefix.
// Passing anything other than a pointer to a struct is guaranteed to end badly.
func Decode(pfx string, v interface{}) error {
	h, err := DefaultHandle()
	if err != nil {
		return err
	}
	return h.Decode(pfx, v)
}

// Decode fetches params into a struct using "ssm" struct tags.
// Nested structs are passed the tag of the parent as a prefix.
// Passing anything other than a pointer to a struct is guaranteed to end badly.
func Encode(pfx string, v interface{}) error {
	h, err := DefaultHandle()
	if err != nil {
		return err
	}
	return h.Encode(pfx, v)
}

type missingErr struct {
	params []*string
}

func (e missingErr) Error() string {
	return fmt.Sprintf("invalid params: %v", e.params)
}

func IsMissing(err error) bool {
	_, ok := err.(missingErr)
	return ok
}
