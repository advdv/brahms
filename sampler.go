package brahms

import (
	"math"
	"math/rand"

	"github.com/dgryski/go-farm"
)

// Sampler holds a sample from a node stream such that it is not biased by the
// nr of times a appears in the stream.
type Sampler struct {
	seeds  []uint64
	mins   []uint64
	sample []Node
}

// NewSampler initializes a sampler with the provided source of randomness
func NewSampler(rnd *rand.Rand, l2 int) (s *Sampler) {
	s = &Sampler{
		mins:   make([]uint64, l2),
		sample: make([]Node, l2),
		seeds:  make([]uint64, l2),
	}

	for i := 0; i < l2; i++ {
		s.mins[i] = math.MaxUint64
		s.seeds[i] = rnd.Uint64()
	}

	return
}

// Update the sampler with a new set of ids
func (s *Sampler) Update(v View) {
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
	v = View{}
	for _, n := range s.sample {
		if n.IsZero() {
			continue
		}

		v[n.Hash()] = n
	}

	return
}
