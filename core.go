package brahms

import (
	"math/rand"
	"sync"
	"time"
)

// Core keeps the state of a node in the gossip network
type Core struct {
	rnd     *rand.Rand
	self    *Node
	view    View
	pushes  chan Node
	params  P
	sampler *Sampler
	tr      Transport
	active  bool
	mu      sync.RWMutex
}

// NewCore initializes the core
func NewCore(rnd *rand.Rand, self *Node, v0 View, p P, tr Transport) (a *Core) {
	a = &Core{
		self:    self,
		view:    v0,
		pushes:  make(chan Node, 100),
		params:  p,
		sampler: NewSampler(rnd, p.L2(), tr),
		tr:      tr,
		rnd:     rnd,
		active:  true,
	}

	//initialize the sampler with our initial view
	a.sampler.Update(v0)
	return
}

// Self returns this core's own info
func (h *Core) Self() (n *Node) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.self
}

// ValidateSample validates if all samples are still responding
func (h *Core) ValidateSample(to time.Duration) {
	h.sampler.Validate(to)
}

// UpdateView runs the algorithm to update the view
func (h *Core) UpdateView(to time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.view = Brahms(h.self, h.rnd, h.params, to, h.sampler, h.tr, h.pushes, h.view)
}

// IsActive is called whenever a remote needs to know if this core is still up
func (h *Core) IsActive() (ok bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.active
}

// ReadView returns a copy of our current local view
func (h *Core) ReadView() View {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.active {
		return View{} //we're no longer active, return nothing
	}

	return h.view.Copy()
}

// ReceiveNode gets called when another peer pushes its info
func (h *Core) ReceiveNode(other Node) {
	select {
	case h.pushes <- other:
	default: //push buffer is full, discard
	}
}

// Deactivate clears the view and sets the core to non-active state
func (h *Core) Deactivate() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.active = false
	h.view = View{}
}

// Sample returns a copy of the peer samples this core has
func (h *Core) Sample() View {
	return h.sampler.Sample()
}
