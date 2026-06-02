//go:build integration

package example_test

import (
	"os"
	"testing"

	"github.com/elloloop/project-scaffold/packages/go/testkit"
)

func TestRepository_Contract_Integration(t *testing.T) {
	if os.Getenv("EXAMPLE_DATABASE_URL") == "" {
		t.Skip("set EXAMPLE_DATABASE_URL to run this against a real repository adapter")
	}
	testkit.RunStoreContract(t, func() testkit.Store {
		return testkit.NewFakeStore(testkit.NewFakeClock(testkit.FixedClockTime))
	})
	t.Log("contract suite is ready; swap the factory for the real adapter when it lands")
}
