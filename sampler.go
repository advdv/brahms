package brahms

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/dgryski/go-farm"
)

// Prober allows for probing peers to determine if they are still online
type Prober interface {
	Probe(ctx context.Context, c chan<- NID, id NID, n Node)
}

// Sampler holds a sample from a node stream such that it is not biased by the
// nr of times a appears in the stream.
type Sampler struct {
	seeds   []uint64
	mins    []uint64
	sample  []Node
	invalid map[NID]time.Time

	ito    time.Duration
	prober Prober
	mu     sync.RWMutex
}

// NewSampler initializes a sampler with the provided source of randomness
func NewSampler(rnd *rand.Rand, l2 int, pr Prober, ito time.Duration) (s *Sampler) {
	s = &Sampler{
		mins:    make([]uint64, l2),
		sample:  make([]Node, l2),
		seeds:   make([]uint64, l2),
		invalid: make(map[NID]time.Time),
		ito:     ito,
		prober:  pr,
	}

	for i := 0; i < l2; i++ {
		s.mins[i] = math.MaxUint64
		s.seeds[i] = rnd.Uint64()
	}

	return
}

// Validate if the currently sampled nodes are still alive
func (s *Sampler) Validate(to time.Duration) {
	sample := s.Sample()
	probes := make(chan NID, len(sample))

	// @TODO probe only an unpredictable subset every call

	func() {
		ctx, cancel := context.WithTimeout(context.Background(), to)
		defer cancel()

		// probe all currently sampled nodes
		for id, n := range sample {
			go s.prober.Probe(ctx, probes, id, n)
		}

		// wait for timeout
		<-ctx.Done()
	}()

	//read the indexes of all probes that returned a response
	alive := map[NID]struct{}{}
DRAIN:
	for {
		select {
		case id := <-probes:
			alive[id] = struct{}{}
		default:
			break DRAIN
		}
	}

	// remove any sample that didn't respond (in time) to the probe
	s.mu.Lock()
	for i, n := range s.sample {
		id := n.Hash()
		if _, ok := sample[id]; !ok {
			continue //node was not probed, keep it for now
		}

		if _, ok := alive[id]; ok {
			continue //this sample replied to the probe, keep it
		}

		// reset the sample otherwise and mark as invalidated
		s.invalid[id] = time.Now()
		s.sample[i] = Node{}
		s.mins[i] = math.MaxUint64
	}

	// clear old invalidated nodes
	for id, t := range s.invalid {
		if time.Now().Sub(t) < s.ito {
			continue //still fresh
		}

		//eviction expired
		delete(s.invalid, id)
	}

	s.mu.Unlock()
}

// Update the sampler with a new set of ids
func (s *Sampler) Update(v View) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, n := range v.Sorted() {

		// @TODO we could use the crypto hash we're already using to
		// also "HashWithSeed" instead of the farm hash.

		id := n.Hash()
		for i, v := range s.mins {
			hv := farm.Hash64WithSeed(id[:], s.seeds[i])
			if hv < v {
				s.mins[i] = hv
				s.sample[i] = n
			}
		}
	}

	return
}

// Sample returns a un-biases sample from all seen nodes
func (s *Sampler) Sample() (v View) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v = View{}
	for _, n := range s.sample {
		if n.IsZero() {
			continue
		}

		v[n.Hash()] = n
	}

	return
}

// Clear the sampler of all samples and mins
func (s *Sampler) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.mins {
		s.mins[i] = math.MaxUint64
	}

	s.sample = make([]Node, len(s.mins))
}

// RecentlyInvalidated returns whether a given node was recently invalidated
// due to a failing probe
func (s *Sampler) RecentlyInvalidated(id NID) (ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok = s.invalid[id]
	return
}
