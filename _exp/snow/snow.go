package snow

// CID represents the identity of a choice
type CID [32]byte

// NilC is the empty choice (no decision)
var NilC = CID{}

// Pref describes the prefence over a set of values
type Pref map[CID]int

// Size returns the nr of keys times the count for each key
func (p Pref) Size() (s int) {
	for _, n := range p {
		s += n
	}
	return
}

// Count returns how high the preference is for a given choice
func (p Pref) Count(id CID) (c int) {
	return p[id]
}

//Querier allows asking neighbours for a sample of at least K preferences
type Querier interface {
	Query(c CID, k int) Pref
}

// Snow allows a network to reach consensus on a set of values using cellular
// Automata logic.
type Snow struct {
	curr CID  // currently preferred choice
	last CID  // last preferred choice
	cnt  int  // confidence in current preferred choice
	cnfd Pref // confidence in other choices
	acc  CID  // currenctly accepted choice

	q  Querier
	k  int
	kα int
	β  int
}

// NewSnow initializes a new snow decision algorithm. k, α, β are configuration
// constants.
func NewSnow(q Querier, k, kα, β int) (s *Snow) {
	s = &Snow{
		cnfd: Pref{},
		q:    q, k: k, kα: kα, β: β,
	}

	return
}

// Query responds to a peer with this nodes current preference
func (s *Snow) Query(c CID) CID {
	if s.curr == NilC {
		s.curr = c
	}

	return s.curr
}

// Decided returns the accepted choice
func (s *Snow) Decided() CID { return s.acc }

// Decide runs a new iteration of the snowball algorithm. It impements snowball
// as defined in figure 3 of the original avalanche paper.
func (s *Snow) Decide() (c CID) {
	if s.acc != NilC {
		return s.acc //already have an accepted value, return it
	}

	// Skip, wait until this node get queries and a peer brings
	// an initial opinion (line 5)
	if s.curr == NilC {
		return
	}

	// query for neighbour preference (line 7)
	// @TODO add some sort of timeout and retry
	p := s.q.Query(s.curr, s.k)
	if p.Size() < s.k {
		return
	}

	// loop over the prefences to change our confidence (line 8)
	for rid := range p {

		// if the count is too low, do nothing (line 9)
		if p.Count(rid) < s.kα {
			continue
		}

		//increment confidence (line 10)
		s.cnfd[rid]++
		if s.cnfd[rid] > s.cnfd[s.curr] {

			// If new confidence becomes larger the currently prefered choice switch
			// the current prefence to the new value (line 11-12)
			s.curr = rid
		}

		// If we saw a new preference, reset the counter. Else check if we're passed
		// the confidence threshold and ready to accept the value.
		if rid != s.last {
			s.last = rid
			s.cnt = 1 //NOTE: we divert from the paper by setting the initial count to 1
		} else {
			s.cnt++
			if s.cnt > s.β {

				// accept the preference and return it
				s.acc = s.curr
				return s.acc
			}
		}
	}

	//nothing got accepted, return the zero value
	return NilC
}
