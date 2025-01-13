package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jmpa-io/paramstore"
)

func main() {

	// setup tracing.
	ctx := context.TODO()

	// setup client.
	c, err := paramstore.New(ctx, paramstore.WithAWSRegion("ap-southeast-2"))
	if err != nil {
		fmt.Printf("failed to setup client: %v\n", err)
		os.Exit(1)
	}

	// read parameter.
	p, err := c.Get(ctx, "/path/to/my/parameter")
	if err != nil {
		fmt.Printf("failed to get parameter: %v", err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", p)
}
