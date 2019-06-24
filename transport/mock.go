package transport

import (
	"context"

	"github.com/advanderveer/brahms"
)

// MockTransport allows mocking of other peers
type MockTransport struct {
	pushed brahms.View
	pulls  map[brahms.NID]brahms.View
}

// NewMockTransport inits a new mock
func NewMockTransport() *MockTransport {
	return &MockTransport{pushed: brahms.View{}, pulls: map[brahms.NID]brahms.View{}}
}

// SetPull imitates a peer responded to a pull
func (t *MockTransport) SetPull(id brahms.NID, v brahms.View) {
	t.pulls[id] = v
}

// DidPush returns whether a peer pushed
func (t *MockTransport) DidPush(id brahms.NID) (ok bool) {
	_, ok = t.pushed[id]
	return
}

// Probe implements probe
func (t *MockTransport) Probe(ctx context.Context, c chan<- int, i int, n brahms.Node) {
	c <- i
}

// Push implements a push
func (t *MockTransport) Push(ctx context.Context, self brahms.Node, to brahms.Node) {
	t.pushed[self.Hash()] = self
}

// Pull implements a pull
func (t *MockTransport) Pull(ctx context.Context, c chan<- brahms.View, from brahms.Node) {
	v, ok := t.pulls[from.Hash()]
	if !ok {
		return
	}

	c <- v
}
