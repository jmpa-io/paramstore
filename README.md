<!-- markdownlint-disable MD041 MD010 -->
<p align="center">
    <img src="docs/logo.png">
</p>

## `paramstore`

```diff
+ 🍱 A Go abstraction over AWS SSM Parameter Store: https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html.
```

## `Usage`

Below is a basic example of how to get started with this package:

```go
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
```

For more explicit examples, see the `cmd/*/main.go` files for details.

## `License`

This work is published under the MIT license.

Please see the [`LICENSE`](./LICENSE) file for details.
