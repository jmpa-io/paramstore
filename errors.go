package paramstore

import (
	"fmt"
)

// Error defines an error returned from this package; used to format errors
// consistently.
type Error struct {
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("ssm: %s", e.Message)
}

// this struct simplifies checking for missing params.
type missingErr struct {
	params []*string
}

func (e missingErr) Error() string {
	return fmt.Sprintf("invalid params: %v", e.params)
}

// IsMissing determines if the error is a missing error type.
func IsMissing(err error) bool {
	_, ok := err.(missingErr)
	return ok
}
