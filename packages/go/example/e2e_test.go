//go:build e2e

package example_test

import (
	"os"
	"testing"
)

func TestNoteFlow_E2E(t *testing.T) {
	base := os.Getenv("EXAMPLE_E2E_BASE_URL")
	if base == "" {
		t.Skip("set EXAMPLE_E2E_BASE_URL to run e2e tests")
	}
	t.Skip("e2e scaffold: wire a real authenticated create-get flow here")
}
