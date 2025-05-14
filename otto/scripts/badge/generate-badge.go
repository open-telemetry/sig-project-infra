// SPDX-License-Identifier: Apache-2.0

// Package badge provides functionality for generating code coverage badges.
package badge

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	// CoverageFile is the path to the coverage output file.
	CoverageFile = flag.String("coverage", "", "Path to the coverage output file")
	// OutputFile is the path to the output SVG file.
	OutputFile = flag.String("output", "coverage-badge.svg", "Path to the output SVG file")
)

// GenerateBadge reads a coverage file and generates an SVG badge.
func GenerateBadge() error {
	if *CoverageFile == "" {
		return errors.New("coverage file path is required")
	}

	// Validate coverage file path (avoid command injection)
	if strings.Contains(*CoverageFile, " ") || strings.Contains(*CoverageFile, ";") ||
		strings.Contains(*CoverageFile, "&") || strings.Contains(*CoverageFile, "|") ||
		strings.Contains(*CoverageFile, ">") || strings.Contains(*CoverageFile, "<") {
		return fmt.Errorf("invalid coverage file path: %s", *CoverageFile)
	}

	// First run go tool cover -func to get formatted coverage
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

	// Generate badge
	badge := GenerateSVG(coverage)

	// Write badge to file - use more restrictive file permission (0o600)
	err = os.WriteFile(*OutputFile, []byte(badge), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write badge file: %w", err)
	}

	log.Printf("Coverage badge generated with %.1f%% coverage\n", coverage)
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

// GenerateSVG creates an SVG badge for the given coverage percentage.
func GenerateSVG(coverage float64) string {
	color := getBadgeColor(coverage)

	// SVG template for the badge
	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="108" height="20" role="img" aria-label="coverage: %.1f%%">
  <title>coverage: %.1f%%</title>
  <linearGradient id="s" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <clipPath id="r">
    <rect width="108" height="20" rx="3" fill="#fff"/>
  </clipPath>
  <g clip-path="url(#r)">
    <rect width="61" height="20" fill="#555"/>
    <rect x="61" width="47" height="20" fill="%s"/>
    <rect width="108" height="20" fill="url(#s)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110">
    <text aria-hidden="true" x="315" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="510">coverage</text>
    <text x="315" y="140" transform="scale(.1)" fill="#fff" textLength="510">coverage</text>
    <text aria-hidden="true" x="835" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="370">%.1f%%</text>
    <text x="835" y="140" transform="scale(.1)" fill="#fff" textLength="370">%.1f%%</text>
  </g>
</svg>`,
		coverage,
		coverage,
		color,
		coverage,
		coverage,
	)
}

// getBadgeColor returns a color based on coverage percentage.
func getBadgeColor(coverage float64) string {
	if coverage >= 90 {
		return "#4c1"
	} else if coverage >= 80 {
		return "#97CA00"
	} else if coverage >= 70 {
		return "#A4A61D"
	} else if coverage >= 60 {
		return "#DFB317"
	} else if coverage >= 50 {
		return "#FE7D37"
	}
	return "#E05D44"
}
