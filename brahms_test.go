package brahms

import (
	"math/rand"
	"testing"
	"time"

	"github.com/advanderveer/go-test"
)

func TestBrahmsNoReply(t *testing.T) {
	p, _ := NewParams(0.1, 0.7, 0.2, 10, 2)
	r := rand.New(rand.NewSource(1))
	s := NewSampler(r, p.l2())
	self := NID{0x01}

	p0 := make(chan NID)
	v0 := NewView(NID{0x01})
	tr0 := NewMockTransport()

	v1 := Brahms(self, r, p, time.Millisecond*10, s, tr0, p0, v0)

	//view should be unchanged transport returned nothing
	test.Equals(t, v0, v1)

	//should have pushed our own nid
	test.Equals(t, true, tr0.DidPush(self))

	//sample should be empty
	test.Equals(t, View{}, s.Sample())
}

func TestBrahmsWithJustPushes(t *testing.T) {
	p, _ := NewParams(0.1, 0.7, 0.2, 10, 2)
	r := rand.New(rand.NewSource(1))
	s := NewSampler(r, p.l2())
	self := NID{0x01}

	id1 := NID{0x01, 0x02}
	p0 := make(chan NID, 10)
	p0 <- id1
	v0 := NewView(NID{0x01})
	tr0 := NewMockTransport()

	// with just a pull response we update the view with just that info
	v1 := Brahms(self, r, p, time.Millisecond*10, s, tr0, p0, v0)
	test.Equals(t, 0, len(p0))
	test.Equals(t, NewView(id1), v1)

	// but the pushed id should have been added to the sample
	test.Equals(t, NewView(id1), s.Sample())

	t.Run("with too many pushes", func(t *testing.T) {
		p1 := make(chan NID, 10)
		p1 <- NID{0xaa, 0xaa}
		p1 <- NID{0xbb, 0xbb} //with the given params this is too much push

		v2 := Brahms(NID{0xff, 0xff}, r, p, time.Millisecond*10, s, tr0, p1, v0)

		//with too many pushes the view shouldn't have changed
		test.Equals(t, v2, v0)
	})
}

func TestBrahmsWithPullsAndPushes(t *testing.T) {
	p, _ := NewParams(0.1, 0.7, 0.2, 10, 2)
	r := rand.New(rand.NewSource(1))
	s := NewSampler(r, p.l2())
	self := NID{0x01}
	other := NID{0x02}

	v0 := NewView(other)
	pull1 := NID{0x01, 0x02}
	push1 := NID{0x02, 0x02}
	p0 := make(chan NID, 10)
	p0 <- push1
	tr0 := NewMockTransport()
	tr0.SetPull(other, NewView(pull1, pull1))

	//with both pushes and pulls the view should get updated
	v1 := Brahms(self, r, p, time.Millisecond*10, s, tr0, p0, v0)
	test.Equals(t, NewView(pull1, push1), v1)
	test.Equals(t, NewView(pull1, push1), s.Sample())
}
