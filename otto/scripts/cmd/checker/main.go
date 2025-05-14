// SPDX-License-Identifier: Apache-2.0

// Command-line utility to check if code coverage meets a threshold
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/open-telemetry/sig-project-infra/otto/scripts/checker"
)

func main() {
	flag.Parse()

	if err := checker.CheckCoverage(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
