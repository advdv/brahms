package transport

import (
	"context"
	"sync"

	"github.com/advanderveer/brahms"
)

// MemNetTransport is an in-memory transport that allows cores to directly
// call each others handlers
type MemNetTransport struct {
	cores map[brahms.NID]*brahms.Core
	mu    sync.RWMutex
}

// NewMemNetTransport inits the new mem transport
func NewMemNetTransport() *MemNetTransport {
	return &MemNetTransport{cores: make(map[brahms.NID]*brahms.Core)}
}

// AddCore adds a core to the network
func (t *MemNetTransport) AddCore(c *brahms.Core) {
	t.mu.Lock()
	defer t.mu.Unlock()

	self := c.Self()
	t.cores[self.Hash()] = c
}

// Probe implements probe
func (t *MemNetTransport) Probe(ctx context.Context, cc chan<- brahms.NID, id brahms.NID, n brahms.Node) {
	t.mu.RLock()
	c, ok := t.cores[n.Hash()]
	if !ok {
		panic("no core known for: " + n.String())
	}

	t.mu.RUnlock()
	if c.IsActive() {
		cc <- id
	}
}

// Push implements a push
func (t *MemNetTransport) Push(ctx context.Context, self brahms.Node, to brahms.Node) {
	t.mu.RLock()
	c, ok := t.cores[to.Hash()]
	if !ok {
		panic("no core known for: " + to.String())
	}

	t.mu.RUnlock()
	c.ReceiveNode(self)
}

// Pull implements a pull
func (t *MemNetTransport) Pull(ctx context.Context, cc chan<- brahms.View, from brahms.Node) {
	t.mu.RLock()
	c, ok := t.cores[from.Hash()]
	if !ok {
		panic("no core known for: " + from.String())
	}

	t.mu.RUnlock()
	cc <- c.ReadView()
}
