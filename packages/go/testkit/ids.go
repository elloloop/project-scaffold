package testkit

import (
	"fmt"
	"sync/atomic"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
)

// SeqIDGen is a deterministic ports.IDGenerator. Ids are "<prefix>-1",
// "<prefix>-2", ... - stable across runs and collision-free within a test, so
// assertions and golden files don't churn on random UUIDs.
type SeqIDGen struct {
	prefix string
	n      atomic.Uint64
}

var _ ports.IDGenerator = (*SeqIDGen)(nil)

// NewSeqIDGen returns a generator that prefixes each id with prefix.
func NewSeqIDGen(prefix string) *SeqIDGen {
	return &SeqIDGen{prefix: prefix}
}

// NewID returns the next sequential id. Safe for concurrent use.
func (g *SeqIDGen) NewID() string {
	return fmt.Sprintf("%s-%d", g.prefix, g.n.Add(1))
}
