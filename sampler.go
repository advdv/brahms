package brahms

import (
	"encoding/hex"
	"math"
	"math/rand"

	"github.com/dgryski/go-farm"
)

// NID is a node id
type NID [32]byte

func (id NID) String() string { return hex.EncodeToString(id[:2]) }

// IsNil returns whether the is its zero value
func (id NID) IsNil() bool {
	return id == NID{}
}

// Sampler holds a sample from a node id stream such that it is not biased by the
// nr of times an id appears in the stream.
type Sampler struct {
	seeds  []uint64
	mins   []uint64
	sample []NID
}

// NewSampler initializes a sampler with the provided source of randomness
func NewSampler(rnd *rand.Rand, l2 int) (s *Sampler) {
	s = &Sampler{
		mins:   make([]uint64, l2),
		sample: make([]NID, l2),
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
	for id := range v {
		for i, v := range s.mins {
			hv := farm.Hash64WithSeed(id[:], s.seeds[i]) + uint64(i)
			if hv < v {
				s.mins[i] = hv
				s.sample[i] = id
			}
		}
	}

	return
}

// Sample returns a un-biases sample from all seen nodes
func (s *Sampler) Sample() (v View) {
	v = View{}
	for _, id := range s.sample {
		if id.IsNil() {
			continue
		}

		v[id] = struct{}{}
	}

	return
}
