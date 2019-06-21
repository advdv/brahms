package brahms

import "context"

// Transport describes how a node communicates with its peers
type Transport interface {
	Push(ctx context.Context, self Node, to Node)
	Pull(ctx context.Context, c chan<- View, from Node)
	Prober
}

// MemNetTransport is an in-memory transport that allows cores to directly
// call each others handlers
type MemNetTransport struct {
	cores map[NID]*Core
}

// NewMemNetTransport inits the new mem transport
func NewMemNetTransport() *MemNetTransport {
	return &MemNetTransport{cores: make(map[NID]*Core)}
}

// AddCore adds a core to the network
func (t *MemNetTransport) AddCore(c *Core) {
	t.cores[c.Self().Hash()] = c
}

// Probe implements probe
func (t *MemNetTransport) Probe(ctx context.Context, cc chan<- int, i int, n Node) {
	c, ok := t.cores[n.Hash()]
	if !ok {
		panic("no core known for: " + n.String())
	}

	if c.HandleProbe() {
		cc <- i
	}
}

// Push implements a push
func (t *MemNetTransport) Push(ctx context.Context, self Node, to Node) {
	c, ok := t.cores[to.Hash()]
	if !ok {
		panic("no core known for: " + to.String())
	}

	c.HandlePush(self)
}

// Pull implements a pull
func (t *MemNetTransport) Pull(ctx context.Context, cc chan<- View, from Node) {
	c, ok := t.cores[from.Hash()]
	if !ok {
		panic("no core known for: " + from.String())
	}

	cc <- c.HandlePull()
}

// MockTransport allows mocking of other peers
type MockTransport struct {
	pushed View
	pulls  map[NID]View
}

// NewMockTransport inits a new mock
func NewMockTransport() *MockTransport {
	return &MockTransport{pushed: View{}, pulls: map[NID]View{}}
}

// SetPull imitates a peer responded to a pull
func (t *MockTransport) SetPull(id NID, v View) {
	t.pulls[id] = v
}

// DidPush returns whether a peer pushed
func (t *MockTransport) DidPush(id NID) (ok bool) {
	_, ok = t.pushed[id]
	return
}

// Probe implements probe
func (t *MockTransport) Probe(ctx context.Context, c chan<- int, i int, n Node) {
	c <- i
}

// Push implements a push
func (t *MockTransport) Push(ctx context.Context, self Node, to Node) {
	t.pushed[self.Hash()] = self
}

// Pull implements a pull
func (t *MockTransport) Pull(ctx context.Context, c chan<- View, from Node) {
	v, ok := t.pulls[from.Hash()]
	if !ok {
		return
	}

	c <- v
}
