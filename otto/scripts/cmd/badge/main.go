// SPDX-License-Identifier: Apache-2.0

// Command-line utility to generate a coverage badge
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/open-telemetry/sig-project-infra/otto/scripts/badge"
)

func main() {
	flag.Parse()

	if err := badge.GenerateBadge(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
