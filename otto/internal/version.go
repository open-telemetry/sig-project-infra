// SPDX-License-Identifier: Apache-2.0

package internal

// Version is set during build time via GoReleaser.
// Source version is managed by release-please-action.
var Version = "dev"

// GetVersion returns the current Otto version.
func GetVersion() string {
	return Version
}
