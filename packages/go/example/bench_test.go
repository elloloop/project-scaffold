package example_test

import (
	"context"
	"testing"

	"github.com/elloloop/project-scaffold/packages/go/example"
	"github.com/elloloop/project-scaffold/packages/go/testkit"
)

func BenchmarkNormalizeTag(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = example.NormalizeTag("High Priority Bug Report")
	}
}

func BenchmarkNoteCreate(b *testing.B) {
	store := testkit.NewFakeStore(testkit.NewFakeClock(testkit.FixedClockTime))
	svc := example.NewNoteService(noteRepo{store: store}, testkit.NewSeqIDGen("n"))
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := svc.Create(ctx, "t1", "user:1", "body"); err != nil {
			b.Fatal(err)
		}
	}
}
