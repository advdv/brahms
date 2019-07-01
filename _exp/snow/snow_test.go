package snow

import (
	"math/rand"
	"testing"

	"github.com/advanderveer/go-test"
)

func TestEmptySnow(t *testing.T) {
	k, kα, β := 5, 3, 10
	var q Querier

	s := NewSnow(q, k, kα, β)

	var c CID
	for i := 0; i < 1000; i++ {
		c = s.Decide()
	}

	// without outside input it should remain undecided
	test.Equals(t, NilC, c)
}

func TestTooSmallQueries(t *testing.T) {
	k, kα, β, q := 5, 3, 10, QuerierFunc(func(c CID, k int) Pref { return Pref{CID{0x01}: 1} })

	s := NewSnow(q, k, kα, β)
	test.Equals(t, CID{0x01}, s.Query(CID{0x01}))
	test.Equals(t, CID{0x01}, s.Query(CID{0x01}))

	var c CID
	for i := 0; i < 1000; i++ {
		c = s.Decide()
	}

	// without query returning anything it should remain undecided event if
	// someone brought in an initial choice if query keeps returning samples
	// that are too small
	test.Equals(t, NilC, c)
}

func TestLargeEnoughQuery(t *testing.T) {
	k, kα, β, c1 := 5, 3, 10, CID{0x01}

	q := QuerierFunc(func(c CID, k int) Pref { return Pref{c1: k} })
	s := NewSnow(q, k, kα, β)
	s.Query(c1)

	var c CID
	for i := 0; i < β; i++ {
		c = s.Decide()
	}

	test.Equals(t, NilC, c) //after β iterations is still undecided
	test.Equals(t, NilC, s.Decided())
	c = s.Decide()
	test.Equals(t, c1, c) //but one more tips into favor
	test.Equals(t, c1, s.Decided())
}

func TestQuerySwitchingOpinion(t *testing.T) {
	k, kα, β, c1, c2, c3 := 5, 3, 10, CID{0x01}, CID{0x02}, CID{0x02}
	qp := Pref{c1: k}

	q := QuerierFunc(func(c CID, k int) Pref { return qp })
	s := NewSnow(q, k, kα, β)
	s.Query(c1)

	var c CID
	for i := 0; i < β; i++ {
		c = s.Decide()
	}

	test.Equals(t, NilC, c) //just, undecided

	//now reduce theo original preference, add some of our own. but not enough
	//to take the new preference into account.
	qp[c1] = 2
	qp[c2] = 2
	for i := 0; i < 100; i++ {
		c = s.Decide()
	}

	//now new preference has enough to take over the original preference
	qp[c2] = k
	for i := 0; i < β; i++ {
		c = s.Decide()
	}

	test.Equals(t, NilC, c) //still undecided
	c = s.Decide()
	test.Equals(t, c2, c) //now, just decided

	// now event if we switch to some other preference it will always keep returning
	// the already accepted choice
	delete(qp, c1)
	delete(qp, c2)
	qp[c3] = k
	for i := 0; i < 100; i++ {
		c = s.Decide()
	}

	test.Equals(t, c2, c) //now, just decided
}

func TestMemoryNetwork(t *testing.T) {
	n, k, kα, β, rounds := 500, 10, 6, 10, 27

	choices := []CID{
		CID{0x01},
		CID{0x02},
		CID{0x03},
		CID{0x04},
	}

	s := int64(5)
	q := NewMemNetQuerier(s)
	r := rand.New(rand.NewSource(s))

	snows := make([]*Snow, 0, n)
	for i := 0; i < n; i++ {
		s := NewSnow(q, k, kα, β)
		s.Query(choices[r.Intn(len(choices))]) //init with random opinion
		q.Add(s)

		snows = append(snows, s)
	}

	for i := 0; i < rounds; i++ {
		for _, s := range snows {
			s.Decide()
		}
	}

	c := snows[0].Decided()
	test.Assert(t, c != NilC, "should have decided something")

	for _, s := range snows {
		test.Assert(t, c == s.Decided(), "all should have decided on the same choice")
	}
}
