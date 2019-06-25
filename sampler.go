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
	Probe(ctx context.Context, c chan<- int, idx int, n Node)
}

// Sampler holds a sample from a node stream such that it is not biased by the
// nr of times a appears in the stream.
type Sampler struct {
	seeds  []uint64
	mins   []uint64
	sample []Node
	prober Prober
	mu     sync.RWMutex
}

// NewSampler initializes a sampler with the provided source of randomness
func NewSampler(rnd *rand.Rand, l2 int, pr Prober) (s *Sampler) {
	s = &Sampler{
		mins:   make([]uint64, l2),
		sample: make([]Node, l2),
		seeds:  make([]uint64, l2),
		prober: pr,
	}

	for i := 0; i < l2; i++ {
		s.mins[i] = math.MaxUint64
		s.seeds[i] = rnd.Uint64()
	}

	return
}

// Validate if the currently sampled nodes are still alive
func (s *Sampler) Validate(to time.Duration) {
	probes := make(chan int, len(s.sample))

	func() {
		ctx, cancel := context.WithTimeout(context.Background(), to)
		defer cancel()

		// probe all currently sampled nodes
		s.mu.RLock()
		for i, n := range s.sample {
			if n.IsZero() {
				continue
			}

			go s.prober.Probe(ctx, probes, i, n)
		}
		s.mu.RUnlock()

		// wait for timeout
		<-ctx.Done()
	}()

	//read the indexes of all probes that returned a response
	alive := map[int]struct{}{}
DRAIN:
	for {
		select {
		case i := <-probes:
			alive[i] = struct{}{}
		default:
			break DRAIN
		}
	}

	s.mu.Lock()
	for i := range s.sample {
		if _, ok := alive[i]; ok {
			continue //this sample replied
		}

		// reset the sample otherwise. NOTE: there is a race condition here since it
		// is possible that the node that is underneath this index has changed to
		// another node at the time the probe returns. This is ok since we wanted to
		// replace it anyway and this will happen soon enough after this delete.
		s.sample[i] = Node{}
		s.mins[i] = math.MaxUint64
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
