package brahms

import (
	"math/rand"
	"time"
)

// Core implements the core algorithm
type Core struct {
	rnd     *rand.Rand
	self    *Node
	view    View
	pushes  chan Node
	params  P
	sampler *Sampler
	tr      Transport
	alive   bool
}

// NewCore initializes the core
func NewCore(rnd *rand.Rand, self *Node, v0 View, p P, tr Transport) (a *Core) {
	a = &Core{
		self:    self,
		view:    v0,
		pushes:  make(chan Node, 100),
		params:  p,
		sampler: NewSampler(rnd, p.l2(), tr),
		tr:      tr,
		rnd:     rnd,
		alive:   true,
	}

	//initialize the sampler with our initial view
	a.sampler.Update(v0)
	return
}

// Self returns this core's own node info
func (h *Core) Self() *Node { return h.self }

// ValidateSample validates if all samples are still responding
func (h *Core) ValidateSample(to time.Duration) {
	h.sampler.Validate(to)
}

// UpdateView runs the algorithm to update the view
func (h *Core) UpdateView(to time.Duration) {
	h.view = Brahms(h.self, h.rnd, h.params, to, h.sampler, h.tr, h.pushes, h.view)
}

// HandleProbe is called whenever a remote needs to know if this core is still up
func (h *Core) HandleProbe() (ok bool) {
	return h.alive
}

// HandlePull responds to pulls by returning a copy of our view
func (h *Core) HandlePull() View {
	if !h.alive {
		return View{} //we're no longer active, return nothing
	}

	return h.view.Copy()
}

// HandlePush handles incoming node info
func (h *Core) HandlePush(other Node) {
	select {
	case h.pushes <- other:
	default: //push buffer is full, discard
	}
}
