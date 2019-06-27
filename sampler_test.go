package brahms_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/advanderveer/brahms"
	"github.com/advanderveer/go-test"
)

type proberFunc func(ctx context.Context, c chan<- brahms.NID, id brahms.NID, n brahms.Node)

func (pr proberFunc) Probe(ctx context.Context, c chan<- brahms.NID, i brahms.NID, n brahms.Node) {
	pr(ctx, c, i, n)
}

func TestSampler(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	pr := proberFunc(func(ctx context.Context, c chan<- brahms.NID, id brahms.NID, n brahms.Node) { c <- id })
	s := brahms.NewSampler(r, 10, pr, time.Second)

	t.Run("empty sampler should return empty view as sample", func(t *testing.T) {
		test.Equals(t, brahms.View{}, s.Sample())
	})

	n1 := brahms.N("127.0.0.1", 1)
	n2 := brahms.N("127.0.0.1", 2)
	n3 := brahms.N("127.0.0.1", 3)
	n4 := brahms.N("127.0.0.1", 4)

	s.Update(brahms.NewView(n1))
	s.Update(brahms.NewView(n2))
	s.Update(brahms.NewView(n3))
	s.Update(brahms.NewView(n4))
	for i := 0; i < 100; i++ {
		s.Update(brahms.NewView(n3))
	}

	s.Update(brahms.NewView(n4))
	for i := 0; i < 1000; i++ {
		s.Update(brahms.NewView(n3))
	}

	test.Equals(t, brahms.NewView(n2, n2, n4, n3), s.Sample())

	t.Run("clearing", func(t *testing.T) {
		s.Clear()
		test.Equals(t, brahms.View{}, s.Sample())
		s.Update(brahms.NewView(n3))
		test.Equals(t, brahms.NewView(n3), s.Sample())
	})
}

func TestSamplerValidation(t *testing.T) {
	n1 := brahms.N("127.0.0.1", 1)
	n2 := brahms.N("127.0.0.1", 2)
	n3 := brahms.N("127.0.0.1", 3)
	n4 := brahms.N("127.0.0.1", 4)

	pr := proberFunc(func(ctx context.Context, c chan<- brahms.NID, id brahms.NID, n brahms.Node) {
		if n.IsZero() {
			t.Fatalf("probe func called with zero node")
		}

		if n.Hash() == n3.Hash() {
			return //n3 doesn't respond
		}

		c <- id
	})

	r := rand.New(rand.NewSource(3))
	s := brahms.NewSampler(r, 15, pr, time.Millisecond*10)
	s.Validate(time.Millisecond)

	s.Update(brahms.NewView(n1, n2, n3, n4))
	test.Equals(t, brahms.NewView(n1, n2, n3, n4), s.Sample())
	test.Equals(t, false, s.RecentlyInvalidated(n3.Hash()))

	s.Validate(time.Millisecond)
	test.Equals(t, brahms.NewView(n1, n2, n4), s.Sample()) //n3 was reset
	test.Equals(t, true, s.RecentlyInvalidated(n3.Hash()))
	test.Equals(t, false, s.RecentlyInvalidated(n1.Hash()))

	s.Validate(time.Millisecond) //should still be recently invaidated
	test.Equals(t, true, s.RecentlyInvalidated(n3.Hash()))

	s.Validate(time.Millisecond * 10) //should expire the invalidation
	test.Equals(t, false, s.RecentlyInvalidated(n3.Hash()))
}
