// SPDX-License-Identifier: Apache-2.0

// Package checker provides functionality for checking code coverage thresholds.
package checker

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	// CoverageFile is the path to the coverage output file.
	CoverageFile = flag.String("coverage", "", "Path to the coverage output file")
	// CoverageThreshold is the minimum required coverage percentage.
	CoverageThreshold = flag.Float64("threshold", 70.0, "Minimum required coverage percentage")
)

// CheckCoverage reads a coverage file and checks if coverage meets threshold.
func CheckCoverage() error {
	if *CoverageFile == "" {
		return errors.New("coverage file path is required")
	}

	// Validate coverage file path (avoid command injection)
	if strings.Contains(*CoverageFile, " ") || strings.Contains(*CoverageFile, ";") ||
		strings.Contains(*CoverageFile, "&") || strings.Contains(*CoverageFile, "|") ||
		strings.Contains(*CoverageFile, ">") || strings.Contains(*CoverageFile, "<") {
		return fmt.Errorf("invalid coverage file path: %s", *CoverageFile)
	}

	// Run go tool cover -func to get formatted coverage
	// #nosec G204 -- CoverageFile is validated above to prevent command injection
	cmd := exec.Command("go", "tool", "cover", "-func", *CoverageFile)
	// Explicitly set no additional input to the command
	cmd.Stdin = nil
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run go tool cover: %w", err)
	}

	// Extract total coverage percentage
	coverage := ExtractCoverage(string(output))

	// Check if coverage meets threshold
	if coverage < *CoverageThreshold {
		return fmt.Errorf("coverage %.1f%% is below the threshold of %.1f%%", coverage, *CoverageThreshold)
	}

	log.Printf("Coverage %.1f%% meets or exceeds the threshold of %.1f%%\n", coverage, *CoverageThreshold)
	return nil
}

// ExtractCoverage extracts the coverage percentage from coverage output.
func ExtractCoverage(coverageData string) float64 {
	// Process the last line of go tool cover -func output
	lines := strings.Split(coverageData, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "total:") {
			re := regexp.MustCompile(`(\d+\.\d+)%`)
			match := re.FindStringSubmatch(lines[i])
			if len(match) > 1 {
				coverage, err := strconv.ParseFloat(match[1], 64)
				if err != nil {
					// Return 0 instead of using log.Fatalf
					return 0
				}
				return coverage
			}
		}
	}

	// Return 0 as a default when coverage can't be extracted
	return 0
}
