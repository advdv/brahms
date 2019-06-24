package transport

import (
	"context"

	"github.com/advanderveer/brahms"
)

// MemNetTransport is an in-memory transport that allows cores to directly
// call each others handlers
type MemNetTransport struct {
	cores map[brahms.NID]*brahms.Core
}

// NewMemNetTransport inits the new mem transport
func NewMemNetTransport() *MemNetTransport {
	return &MemNetTransport{cores: make(map[brahms.NID]*brahms.Core)}
}

// AddCore adds a core to the network
func (t *MemNetTransport) AddCore(c *brahms.Core) {
	t.cores[c.Self().Hash()] = c
}

// Probe implements probe
func (t *MemNetTransport) Probe(ctx context.Context, cc chan<- int, i int, n brahms.Node) {
	c, ok := t.cores[n.Hash()]
	if !ok {
		panic("no core known for: " + n.String())
	}

	if c.IsActive() {
		cc <- i
	}
}

// Push implements a push
func (t *MemNetTransport) Push(ctx context.Context, self brahms.Node, to brahms.Node) {
	c, ok := t.cores[to.Hash()]
	if !ok {
		panic("no core known for: " + to.String())
	}

	c.ReceiveNode(self)
}

// Pull implements a pull
func (t *MemNetTransport) Pull(ctx context.Context, cc chan<- brahms.View, from brahms.Node) {
	c, ok := t.cores[from.Hash()]
	if !ok {
		panic("no core known for: " + from.String())
	}

	cc <- c.ReadView()
}
