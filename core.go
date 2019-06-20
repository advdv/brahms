package brahms

import (
	"math/rand"
	"time"
)

// Core implements the core algorithm
type Core struct {
	rnd     *rand.Rand
	self    NID
	view    View
	pushes  chan NID
	params  P
	sampler *Sampler
	tr      Transport
}

// NewCore initializes the core
func NewCore(rnd *rand.Rand, self NID, v0 View, p P, tr Transport) (a *Core) {
	a = &Core{
		self:    self,
		view:    v0,
		pushes:  make(chan NID, 100),
		params:  p,
		sampler: NewSampler(rnd, p.l2()),
		tr:      tr,
		rnd:     rnd,
	}

	//initialize the sampler with our initial view
	a.sampler.Update(v0)
	return
}

// @TODO probe for sample validation

// ID returns this core's id
func (h *Core) ID() NID { return h.self }

// UpdateView runs the algorithm to update the view
func (h *Core) UpdateView(to time.Duration) {
	h.view = Brahms(h.self, h.rnd, h.params, to, h.sampler, h.tr, h.pushes, h.view)
}

// HandlePull responds to pulls by returning a copy of our view
func (h *Core) HandlePull() View {
	return h.view.Copy()
}

// HandlePush handles incoming ids
func (h *Core) HandlePush(id NID) {
	select {
	case h.pushes <- id:
	default: //push buffer is full, discard
	}
}
