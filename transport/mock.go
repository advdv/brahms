package transport

import (
	"context"
	"sync"

	"github.com/advanderveer/brahms"
)

// MockTransport allows mocking of other peers
type MockTransport struct {
	pushed brahms.View
	pulls  map[brahms.NID]brahms.View
	mu     sync.RWMutex
}

// NewMockTransport inits a new mock
func NewMockTransport() *MockTransport {
	return &MockTransport{pushed: brahms.View{}, pulls: map[brahms.NID]brahms.View{}}
}

// SetPull imitates a peer responded to a pull
func (t *MockTransport) SetPull(id brahms.NID, v brahms.View) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pulls[id] = v
}

// DidPush returns whether a peer pushed
func (t *MockTransport) DidPush(id brahms.NID) (ok bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, ok = t.pushed[id]
	return
}

// Probe implements probe
func (t *MockTransport) Probe(ctx context.Context, c chan<- brahms.NID, id brahms.NID, n brahms.Node) {
	c <- id
}

// Push implements a push
func (t *MockTransport) Push(ctx context.Context, self brahms.Node, to brahms.Node) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pushed[self.Hash()] = self
}

// Pull implements a pull
func (t *MockTransport) Pull(ctx context.Context, c chan<- brahms.View, from brahms.Node) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	v, ok := t.pulls[from.Hash()]
	if !ok {
		return
	}

	c <- v
}

// Emit implements custom message emit
func (t *MockTransport) Emit(ctx context.Context, c chan<- brahms.NID, id brahms.NID, msg []byte, to brahms.Node) {
	panic("not implemented")
}
