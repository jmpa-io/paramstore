<!-- markdownlint-disable MD041 MD010 -->
<p align="center">
    <img src="docs/logo.png">
</p>

## `paramstore`

```diff
+ 🍱 A Go abstraction over AWS SSM Parameter Store: https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html.
```
<a href="LICENSE" target="_blank"><img src="https://img.shields.io/github/license/jmpa-io/paramstore.svg" alt="GitHub License"></a>
[![CI/CD](https://github.com/jmpa-io/paramstore/actions/workflows/cicd.yml/badge.svg)](https://github.com/jmpa-io/paramstore/actions/workflows/cicd.yml)
[![Codecov](https://codecov.io/github/jmpa-io/paramstore/graph/badge.svg)](https://codecov.io/github/jmpa-io/paramstore)

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

## `Tracing`

Every client call (`Get`, `GetMultiple`, `Put`, `Delete`) creates a span via the global OpenTelemetry tracer provider, and failures are recorded on the span. This package only depends on the OTel API: without a tracer provider configured, spans are no-ops. See `cmd/tracing` for a full example, configured with the standard `OTEL_EXPORTER_OTLP_*` environment variables, eg:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run ./cmd/tracing
```

For more explicit examples, see the `cmd/*/main.go` files for details.

## `License`

This work is published under the MIT license.

Please see the [`LICENSE`](./LICENSE) file for details.
