package brahms

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/advanderveer/go-test"
)

func TestSampler(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	pr := proberFunc(func(ctx context.Context, c chan<- int, i int, n Node) { c <- i })
	s := NewSampler(r, 10, pr)

	t.Run("empty sampler should return empty view as sample", func(t *testing.T) {
		test.Equals(t, View{}, s.Sample())
	})

	n1 := N("127.0.0.1", 1)
	n2 := N("127.0.0.1", 2)
	n3 := N("127.0.0.1", 3)
	n4 := N("127.0.0.1", 4)

	s.Update(NewView(n1))
	s.Update(NewView(n2))
	s.Update(NewView(n3))
	s.Update(NewView(n4))
	for i := 0; i < 100; i++ {
		s.Update(NewView(n3))
	}

	s.Update(NewView(n4))
	for i := 0; i < 1000; i++ {
		s.Update(NewView(n3))
	}

	test.Equals(t, NewView(n2, n2, n4, n3), s.Sample())
}

func TestSamplerValidation(t *testing.T) {
	n1 := N("127.0.0.1", 1)
	n2 := N("127.0.0.1", 2)
	n3 := N("127.0.0.1", 3)
	n4 := N("127.0.0.1", 4)

	pr := proberFunc(func(ctx context.Context, c chan<- int, i int, n Node) {
		if n.IsZero() {
			t.Fatalf("probe func called with zero node")
		}

		if n.Hash() == n3.Hash() {
			return //n3 doesn't respond
		}

		c <- i
	})

	r := rand.New(rand.NewSource(3))
	s := NewSampler(r, 15, pr)
	s.Validate(time.Millisecond)

	s.Update(NewView(n1, n2, n3, n4))
	test.Equals(t, NewView(n1, n2, n3, n4), s.Sample())

	s.Validate(time.Millisecond)
	test.Equals(t, NewView(n1, n2, n4), s.Sample()) //n3 was reset
}
