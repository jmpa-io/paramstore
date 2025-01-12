package paramstore

import "fmt"

// ErrClientFailedToSetOption is returned when an option encounters an error
// when trying to be set with the client.
type ErrClientFailedToSetOption struct {
	err error
}

func (e ErrClientFailedToSetOption) Error() string {
	return fmt.Sprintf("failed to set option in client: %v", e.err)
}

type ErrClientFailedToLoadAWSConfig struct {
	err error
}

// ErrClientFailedToLoadAWSConfig is returned when the AWS config isn't able
// to be loaded from the AWS SDK being used in the client. This may occur when
// the environment isn't configured correctly to use the `awscli` for example.
func (e ErrClientFailedToLoadAWSConfig) Error() string {
	return fmt.Sprintf("failed to load AWS config: %v", e.err)
}
