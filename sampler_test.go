package brahms

import (
	"math/rand"
	"testing"

	"github.com/advanderveer/go-test"
)

func TestPRF(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	s := NewSampler(r, 10)

	t.Run("empty sampler should return empty view as sample", func(t *testing.T) {
		test.Equals(t, View{}, s.Sample())
	})

	s.Update(NewView(NID{0x02, 0x04}))
	s.Update(NewView(NID{0x01, 0x04}))
	s.Update(NewView(NID{0x66, 0x66}))
	s.Update(NewView(NID{0x05, 0x04}))
	for i := 0; i < 100; i++ {
		s.Update(NewView(NID{0x66, 0x66}))
	}

	s.Update(NewView(NID{0x05, 0x04}))
	for i := 0; i < 1000; i++ {
		s.Update(NewView(NID{0x66, 0x66}))
	}

	test.Equals(t, NewView(NID{0x01, 0x04}, NID{0x02, 0x04}, NID{0x05, 0x04}, NID{0x66, 0x66}), s.Sample())
}
