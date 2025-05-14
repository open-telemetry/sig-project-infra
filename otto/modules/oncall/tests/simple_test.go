// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSimple is a simple test to make sure the test setup works properly.
func TestSimple(t *testing.T) {
	// This is a more meaningful assertion that tests the Go runtime works correctly
	assert.Equal(t, 2, 1+1, "Basic addition should work")
}
