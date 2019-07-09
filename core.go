package brahms

import (
	"math/rand"
	"net"
	"sync/atomic"
	"time"
)

// Core keeps the state of a node in the gossip network
type Core struct {
	rnd     *rand.Rand
	self    *Node
	view    atomic.Value
	pushes  chan Node
	params  P
	sampler *Sampler
	tr      Transport
	active  int32
}

// NewCore initializes the core
func NewCore(rnd *rand.Rand, self *Node, v0 View, p P, tr Transport, ito time.Duration) (c *Core) {
	c = &Core{
		self:    self,
		pushes:  make(chan Node, p.L1α()+10), // slightly larger then what the algorithm accepts
		params:  p,
		sampler: NewSampler(rnd, p.L2(), tr, ito),
		tr:      tr,
		rnd:     rnd,

		// the active flag is implemented as an atomic uint32 so it can be read
		// concurrently without locking the whole core. This happens when many
		// peers probe this node.
		active: 1,
	}

	// the view is an atomic value such that we have no lock contention
	// on reading and updating it
	c.view.Store(v0)

	//initialize the sampler with our initial view
	c.sampler.Update(v0)
	return
}

// Self returns this core's own info
func (c *Core) Self() (n Node) {
	n.IP = make(net.IP, len(c.self.IP))
	copy(n.IP, c.self.IP)
	n.Port = c.self.Port
	return
}

// ValidateSample validates if all samples are still responding
func (c *Core) ValidateSample(to time.Duration) {
	c.sampler.Validate(c.rnd, c.params.VN(), to)
}

// UpdateView runs the algorithm to update the view
func (c *Core) UpdateView(to time.Duration) {
	c.view.Store(
		Brahms(c.self, c.rnd, c.params, to, c.sampler, c.tr, c.pushes, c.view.Load().(View)),
	)
}

// ReadView returns a copy of our current local view
func (c *Core) ReadView() View {
	if atomic.LoadInt32(&(c.active)) != 1 {
		return View{} //we're no longer active, return nothing
	}

	return c.view.Load().(View).Copy()
}

// IsActive is called whenever a remote needs to know if this core is still up
func (c *Core) IsActive() (ok bool) {
	return atomic.LoadInt32(&(c.active)) == 1
}

// ReceiveNode gets called when another peer pushes its info
func (c *Core) ReceiveNode(other Node) {
	select {
	case c.pushes <- other:
	default: //push buffer is full, discard
	}
}

// Deactivate clears the view and sets the core to non-active state
func (c *Core) Deactivate() {
	atomic.StoreInt32(&(c.active), 0)
	c.view.Store(View{})
	c.sampler.Clear()
}

// Sample returns a copy of the peer samples this core has
func (c *Core) Sample() View {
	return c.sampler.Sample()
}
