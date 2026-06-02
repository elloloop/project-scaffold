package testkit

import "testing"

// FakeStore must satisfy the Store contract - the same suite that a real database
// or Postgres adapter will be run against under -tags=integration.
func TestFakeStore_Contract(t *testing.T) {
	t.Parallel()
	RunStoreContract(t, func() Store {
		return NewFakeStore(NewFakeClock(FixedClockTime))
	})
}
