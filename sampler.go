package brahms

import (
	"context"
	"crypto/sha256"
	"math/big"
	"math/rand"
	"sync"
	"time"
)

// SampleRank describes a rank based on 32 bytes of data as a big number
type SampleRank [32]byte

// ToInt converts the bytes to the big nr
func (sr SampleRank) ToInt() *big.Int {
	return new(big.Int).SetBytes(sr[:])
}

// MaxSampleRank is the maximum rank a sample can reach
var MaxSampleRank = SampleRank{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
}

// Prober allows for probing peers to determine if they are still online
type Prober interface {
	Probe(ctx context.Context, c chan<- NID, id NID, n Node)
}

// Sampler holds a sample from a node stream such that it is not biased by the
// nr of times a appears in the stream.
type Sampler struct {
	seeds   [][32]byte
	mins    []SampleRank
	sample  []Node
	invalid map[NID]time.Time

	ito    time.Duration
	prober Prober
	mu     sync.RWMutex
}

// NewSampler initializes a sampler with the provided source of randomness
func NewSampler(rnd *rand.Rand, l2 int, pr Prober, ito time.Duration) (s *Sampler) {
	s = &Sampler{
		mins:    make([]SampleRank, l2),
		sample:  make([]Node, l2),
		seeds:   make([][32]byte, l2),
		invalid: make(map[NID]time.Time),
		ito:     ito,
		prober:  pr,
	}

	for i := 0; i < l2; i++ {
		s.mins[i] = MaxSampleRank
		rnd.Read(s.seeds[i][:])
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

		// @TODO return early if all probed nodes return in time
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
		s.mins[i] = MaxSampleRank
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

		id := n.Hash()
		for i, v := range s.mins {

			// we use a seeded crypto hash to rank a sample
			hv := SampleRank(sha256.Sum256(append(id[:], s.seeds[i][:]...)))
			if hv.ToInt().Cmp(v.ToInt()) < 0 {
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
		s.mins[i] = MaxSampleRank
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
