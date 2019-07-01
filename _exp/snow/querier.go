package snow

import "math/rand"

// QuerierFunc implements the querier with just a function
type QuerierFunc func(c CID, k int) Pref

// Query neighbours for at least a k size preference
func (qf QuerierFunc) Query(c CID, k int) Pref { return qf(c, k) }

// MemNetQuerier implement querying from in-memory snow instances
type MemNetQuerier struct {
	net map[*Snow]struct{}
	rnd *rand.Rand
}

// NewMemNetQuerier initiations the querier
func NewMemNetQuerier(seed int64) (q *MemNetQuerier) {
	q = &MemNetQuerier{
		net: make(map[*Snow]struct{}),
		rnd: rand.New(rand.NewSource(seed)),
	}
	return
}

// Add a snow instance
func (q *MemNetQuerier) Add(s *Snow) {
	q.net[s] = struct{}{}
}

// Query the memory network
func (q *MemNetQuerier) Query(c CID, k int) (p Pref) {
	ns := make([]*Snow, 0, len(q.net))
	for s := range q.net {
		ns = append(ns, s)
	}

	q.rnd.Shuffle(len(ns), func(i int, j int) {
		ns[i], ns[j] = ns[j], ns[i]
	})

	p = Pref{}
	for i := 0; i < k; i++ {
		if i >= len(ns) {
			break
		}

		p[ns[i].Query(c)]++
	}

	return
}
