package testkit

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// updateGolden is wired to `go test -update`, which rewrites golden files
// instead of comparing against them.
var updateGolden = flag.Bool("update", false, "update golden files in testdata/")

// Golden compares got against the golden file testdata/<name>.golden. Run
// `go test -update` to (re)write the file after an intentional change. Use it
// for stable serialized output (proto JSON, rendered templates, fixtures).
func Golden(t testing.TB, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("testkit: mkdir testdata: %v", err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatalf("testkit: write golden %s: %v", path, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("testkit: read golden %s: %v (run `go test -update` to create it)", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("golden mismatch for %s:\n got:  %s\n want: %s\n(run `go test -update` if the change is intended)", name, got, want)
	}
}
